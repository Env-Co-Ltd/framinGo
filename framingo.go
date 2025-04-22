package framinGo

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/CloudyKit/jet/v6"
	"github.com/Env-Co-Ltd/framinGo/cache"
	"github.com/Env-Co-Ltd/framinGo/filesystems/miniofilesystem"
	"github.com/Env-Co-Ltd/framinGo/filesystems/s3filesystem"
	"github.com/Env-Co-Ltd/framinGo/filesystems/sftpfilesystem"
	"github.com/Env-Co-Ltd/framinGo/filesystems/webdavfilesystem"
	"github.com/Env-Co-Ltd/framinGo/mailer"
	"github.com/Env-Co-Ltd/framinGo/render"
	"github.com/Env-Co-Ltd/framinGo/session"
	"github.com/alexedwards/scs/v2"
	"github.com/dgraph-io/badger/v3"
	"github.com/go-chi/chi/v5"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

const version = "1.0.0"

var myRedisCache *cache.RedisCache
var myBadgerCache *cache.BadgerCache
var redisPool *redis.Pool
var badgerConn *badger.DB

var maintenanceMode bool

// Celeritas is the overall type for the Celeritas package. Members that are exported in this type
// are available to any application that uses it.
type FraminGo struct {
	AppName     string
	Debug       bool
	Version     string
	ErrorLog    *log.Logger
	InfoLog     *log.Logger
	RootPath    string
	Routes      *chi.Mux
	Render      *render.Render
	Session     *scs.SessionManager
	DB          Database
	JetViews    *jet.Set
	config      config
	Encryption  string
	Cache       cache.Cache
	Scheduler   *cron.Cron
	Mail        *mailer.Mail
	Server      Server
	FileSystems map[string]any
	S3          s3filesystem.S3
	Minio       miniofilesystem.Minio
	SFTP        sftpfilesystem.SFTP
	WebDAV      webdavfilesystem.WebDAV
}

type Server struct {
	ServerName string
	Port       string
	Secure     bool
	URL        string
}

type config struct {
	port        string
	renderer    string
	cookie      cookieConfig
	sessionType string
	database    databaseConfig
	redis       redisConfig
	uploads     UploadConfig
}

type UploadConfig struct {
	allowedMineTypes []string
	maxUploadSize    int64
}

// New reads the .env file, creates our application config, populates the Celeritas type with settings
// based on .env values, and creates necessary folders and files if they don't exist
func (f *FraminGo) New(rootPath string) error {
	pathConfig := initPaths{
		rootPath:    rootPath,
		folderNames: []string{"handlers", "migrations", "views", "mail", "data", "public", "tmp", "logs", "middleware", "screenshots"},
	}

	err := f.Init(pathConfig)
	if err != nil {
		return err
	}

	err = f.checkDotEnv(rootPath)
	if err != nil {
		return err
	}

	// read .env
	err = godotenv.Load(rootPath + "/.env")
	if err != nil {
		return err
	}

	// create loggers
	infoLog, errorLog := f.startLoggers()

	// connect to database
	if os.Getenv("DATABASE_TYPE") != "" {
		db, err := f.OpenDB(os.Getenv("DATABASE_TYPE"), f.BuildDSN())
		if err != nil {
			errorLog.Println(err)
			os.Exit(1)
		}
		f.DB = Database{
			DataType: os.Getenv("DATABASE_TYPE"),
			Pool:     db,
		}
	}

	scheduler := cron.New()
	f.Scheduler = scheduler

	if os.Getenv("CACHE") == "redis" || os.Getenv("SESSION_TYPE") == "redis" {
		myRedisCache = f.createClientRedisCache()
		f.Cache = myRedisCache
		redisPool = myRedisCache.Conn
	}

	if os.Getenv("CACHE") == "badger" {
		myBadgerCache = f.createClientBadgerCache()
		f.Cache = myBadgerCache
		badgerConn = myBadgerCache.Conn
		_, err = f.Scheduler.AddFunc("@every 1h", func() {
			_ = myBadgerCache.Conn.RunValueLogGC(0.7)
		})
		if err != nil {
			return err
		}

		f.Scheduler.Start()
	}

	f.InfoLog = infoLog
	f.ErrorLog = errorLog
	f.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
	f.Version = version
	f.RootPath = rootPath
	f.Mail = f.createMailer()
	f.Routes = f.routes().(*chi.Mux)
	// upload config
	exploded := strings.Split(os.Getenv("ALLOWED_FILETYPES"), ",")
	var mineTypes []string
	for _, m := range exploded {
		mineTypes = append(mineTypes, m)
	}
	var maxUploadSize int64
	if max, err := strconv.ParseInt(os.Getenv("MAX_UPLOAD_SIZE"), 10, 64); err != nil {
		maxUploadSize = 10 << 20
	} else {
		maxUploadSize = int64(max)
	}

	f.config = config{
		port:     os.Getenv("PORT"),
		renderer: os.Getenv("RENDERER"),
		cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSISTS"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		sessionType: os.Getenv("SESSION_TYPE"),
		database: databaseConfig{
			database: os.Getenv("DATABASE_TYPE"),
			dsn:      f.BuildDSN(),
		},
		redis: redisConfig{
			Host:     os.Getenv("REDIS_HOST"),
			Password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		},
		uploads: UploadConfig{
			allowedMineTypes: mineTypes,
			maxUploadSize:    maxUploadSize,
		},
	}

	secure := true
	if strings.ToLower(os.Getenv("SECURE")) == "false" {
		secure = false
	}

	f.Server = Server{
		ServerName: os.Getenv("SERVER_NAME"),
		Port:       os.Getenv("PORT"),
		Secure:     secure,
		URL:        os.Getenv("APP_URL"),
	}

	// create session
	sess := session.Session{
		CookieLifetime: f.config.cookie.lifetime,
		CookiePersist:  f.config.cookie.persist,
		CookieName:     f.config.cookie.name,
		SessionType:    f.config.sessionType,
		CookieDomain:   f.config.cookie.domain,
	}

	switch f.config.sessionType {
	case "redis":
		sess.RedisPool = myRedisCache.Conn
	case "mysql", "mariadb", "postgres", "postgresql":
		sess.DBPool = f.DB.Pool
	}

	f.Session = sess.InitSession()
	f.Encryption = os.Getenv("ENCRYPTION_KEY")

	if f.Debug {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
			jet.InDevelopmentMode(),
		)

		f.JetViews = views
	} else {
		var views = jet.NewSet(
			jet.NewOSFileSystemLoader(fmt.Sprintf("%s/views", rootPath)),
		)

		f.JetViews = views

	}

	f.createRenderer()
	f.FileSystems = f.createFileSystems()

	// メールリスナーの起動（nilチェック付き）
	if f.Mail != nil && f.Mail.Jobs != nil && f.Mail.Results != nil {
		go f.Mail.ListenForMail()
	}

	return nil
}

// Init creates necessary folders for our Celeritas application
func (f *FraminGo) Init(p initPaths) error {
	root := p.rootPath
	for _, path := range p.folderNames {
		// create folder if it doesn't exist
		err := f.CreateDirIfNotExist(root + "/" + path)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FraminGo) checkDotEnv(path string) error {
	err := f.CreateFileIfNotExists(fmt.Sprintf("%s/.env", path))
	if err != nil {
		return err
	}
	return nil
}

func (f *FraminGo) startLoggers() (*log.Logger, *log.Logger) {
	var infoLog *log.Logger
	var errorLog *log.Logger

	infoLog = log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog = log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	return infoLog, errorLog
}

func (f *FraminGo) createRenderer() {
	myRenderer := render.Render{
		Renderer: f.config.renderer,
		RootPath: f.RootPath,
		Port:     f.config.port,
		JetViews: f.JetViews,
		Session:  f.Session,
	}
	f.Render = &myRenderer
}

// Mailerを作成する
func (f *FraminGo) createMailer() *mailer.Mail {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	m := mailer.Mail{
		Domain:      os.Getenv("MAIL_DOMAIN"),
		Templates:   f.RootPath + "/mail",
		Host:        os.Getenv("SMTP_HOST"),
		Port:        port,
		Username:    os.Getenv("SMTP_USERNAME"),
		Password:    os.Getenv("SMTP_PASSWORD"),
		Encryption:  os.Getenv("SMTP_ENCRYPTION"),
		FromName:    os.Getenv("FROM_NAME"),
		FromAddress: os.Getenv("FROM_ADDRESS"),
		Jobs:        make(chan mailer.Message, 20),
		Results:     make(chan mailer.Result, 20),
		API:         os.Getenv("MAILER_API"),
		APIKey:      os.Getenv("MAILER_KEY"),
		APIUrl:      os.Getenv("MAILER_URL"),
	}
	return &m
}

// RedisCacheを作成する
func (f *FraminGo) createClientRedisCache() *cache.RedisCache {
	cacheClient := cache.RedisCache{
		Conn:   f.createRedisPool(),
		Prefix: f.config.redis.prefix,
	}

	return &cacheClient
}

// BadgerCacheを作成する
func (f *FraminGo) createClientBadgerCache() *cache.BadgerCache {
	cacheClient := cache.BadgerCache{
		Conn:   f.createBadgerConn(),
		Prefix: f.config.redis.prefix,
	}

	return &cacheClient
}

func (f *FraminGo) createRedisPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50,
		MaxActive:   10000,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp",
				f.config.redis.Host,
				redis.DialPassword(f.config.redis.Password),
			)
		},

		TestOnBorrow: func(conn redis.Conn, t time.Time) error {
			_, err := conn.Do("PING")
			return err
		},
	}
}

func (f *FraminGo) createBadgerConn() *badger.DB {
	db, err := badger.Open(badger.DefaultOptions(f.RootPath + "/tmp/badger"))
	if err != nil {
		return nil
	}
	return db
}

// BuildDSN builds the datasource name for our database, and returns it as a string
func (f *FraminGo) BuildDSN() string {
	var dsn string

	switch os.Getenv("DATABASE_TYPE") {
	case "postgres", "postgresql":
		dsn = fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=%s timezone=UTC connect_timeout=5",
			os.Getenv("DATABASE_HOST"),
			os.Getenv("DATABASE_PORT"),
			os.Getenv("DATABASE_USER"),
			os.Getenv("DATABASE_NAME"),
			os.Getenv("DATABASE_SSL_MODE"))

		// we check to see if a database passsword has been supplied, since including "password=" with nothing
		// after it sometimes causes postgres to fail to allow a connection.
		if os.Getenv("DATABASE_PASS") != "" {
			dsn = fmt.Sprintf("%s password=%s", dsn, os.Getenv("DATABASE_PASS"))
		}

	default:

	}

	return dsn
}

func (f *FraminGo) createFileSystems() map[string]any {
	fileSystems := make(map[string]any)
	if os.Getenv("S3_KEY") != "" {
		s3 := s3filesystem.S3{
			Key:      os.Getenv("S3_KEY"),
			Secret:   os.Getenv("S3_SECRET"),
			Region:   os.Getenv("S3_REGION"),
			Endpoint: os.Getenv("S3_ENDPOINT"),
			Bucket:   os.Getenv("S3_BUCKET"),
		}
		fileSystems["S3"] = s3
		f.S3 = s3
	}
	if os.Getenv("MINIO_SECRET") != "" {
		useSSL := false
		if os.Getenv("MINIO_USESSL") == "true" {
			useSSL = true
		}
		minio := miniofilesystem.Minio{
			Endpoint: os.Getenv("MINIO_ENDPOINT"),
			Key:      os.Getenv("MINIO_KEY"),
			Secret:   os.Getenv("MINIO_SECRET"),
			Region:   os.Getenv("MINIO_REGION"),
			UseSSL:   useSSL,
			Bucket:   os.Getenv("MINIO_BUCKET"),
		}
		fileSystems["MINIO"] = minio
		f.Minio = minio
	}

	if os.Getenv("SFTP_HOST") != "" {
		sftp := sftpfilesystem.SFTP{
			Host: os.Getenv("SFTP_HOST"),
			User: os.Getenv("SFTP_USER"),
			Pass: os.Getenv("SFTP_PASS"),
			Port: os.Getenv("SFTP_PORT"),
		}
		fileSystems["SFTP"] = sftp
		f.SFTP = sftp
	}

	if os.Getenv("WEBDAV_HOST") != "" {
		webDav := webdavfilesystem.WebDAV{
			URL:  os.Getenv("WEBDAV_HOST"),
			User: os.Getenv("WEBDAV_USER"),
			Pass: os.Getenv("WEBDAV_PASS"),
		}
		fileSystems["WEBDAV"] = webDav
		f.WebDAV = webDav
	}

	return fileSystems
}

type RPCServer struct{}

func (r *RPCServer) MaintenanceMode(inMaintenanceMode bool, resp *string) error {
	if inMaintenanceMode {
		maintenanceMode = true
		*resp = "Server is in maintenance mode"
	} else {
		maintenanceMode = false
		*resp = "Server is running normally"
	}
	return nil
}

func (f *FraminGo) ListenRPC() {
	if os.Getenv("RCP_PORT") != "" {
		f.InfoLog.Println("Starting RCP server on port", os.Getenv("RCP_PORT"))
		err := rpc.Register(new(RPCServer))
		if err != nil {
			f.ErrorLog.Println(err)
			return
		}
		listen, err := net.Listen("tcp", "127.0.0.1:"+os.Getenv("RCP_PORT"))
		if err != nil {
			f.ErrorLog.Println(err)
			return
		}
		for {
			rcpConn, err := listen.Accept()
			if err != nil {
				f.ErrorLog.Println(err)
				continue
			}
			go rpc.ServeConn(rcpConn)
		}
	}
}
