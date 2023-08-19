package main

import (
  "io"
  "os"

  "archive/tar"
  "compress/gzip"
  
  //tea "github.com/charmbracelet/bubbletea"
  "github.com/charmbracelet/log"
)

func main() {
  err := makeArchive("files")
  if err != nil {
    log.Fatal("Failed to create archive.", "err", err)
  }
}

/*
 * makeArchive
 * Creates a new archive.
 *
 * Takes the name of the directory to archive.
 */
func makeArchive(dir string) error {
  entries, err := os.ReadDir(dir)
  if err != nil {
    return err
  }
  
  buf, err := os.Create(dir + ".tar.gz")
  if err != nil {
    return err
  }

  gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

  for _, file := range entries {
    filename := e.Name()

   	// Open the file
    file, err := os.Open(filename)
    if err != nil {
   		return err
	  }
  	defer file.Close()
  
	  // Stat the file to get info about it
  	info, err := file.Stat()
	  if err != nil {
		  return err
  	}

	  // Create a tar Header
  	header, err := tar.FileInfoHeader(info, info.Name())
	  if err != nil {
  		return err
	  }

	  header.Name = filename

  	// Write file header to the tar archive
	  err = tw.WriteHeader(header)
  	if err != nil {
  		return err
	  }
  
	  // Copy file content to tar archive
  	_, err = io.Copy(tw, file)
	  if err != nil {
		  return err
  	}
	}

  return nil
}
