package main

import (
	"embed"
	"errors"
	"os"
)

//go:embed templates/*
var templateFS embed.FS

func copyFileFromTemplate(templatePath, targetFile string) error {
  //check if file doed not already exists
  if fileExists(targetFile) {
    return errors.New(targetFile + " already exists")
  }
  
  data,err := templateFS.ReadFile(templatePath)
  if err != nil {
    exitGracefully(err)
  }

  err = copyDataToFile(data, targetFile)
  if err != nil {
    exitGracefully(err)
  }

  return nil
}

func copyDataToFile(data []byte, to string) error {
  err := os.WriteFile(to, data, 0644)
  if err != nil {
    return err
  }

  return nil
}

func fileExists(fileToCheck string) bool {
  if _, err := os.Stat(fileToCheck); os.IsNotExist(err) {
    return false
  }
  return true
}