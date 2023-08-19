/*
 * Hey there, thank you for reading the code to make sure that it's safe
 * and that you can trust it! If you see anything alarming, please open
 * an discussion and let me know, alternatively (if you don't want to
 * create an account) you can email me at beta@hai.haus.
 *
 * ~ Daniel
 */

package main

import (
  "io/ioutil"
  "io"
  "os"

  "archive/tar"
  "compress/gzip"
  
  "crypto/aes"
  "crypto/rand"
  "crypto/cipher"
  
  //tea "github.com/charmbracelet/bubbletea"
  "github.com/charmbracelet/log"
)

func main() {
  err := makeArchive("files")
  if err != nil {
    log.Fatal("Failed to create archive.", "err", err)
  }

  err = encryptFile("files.tar.gz")
  if err != nil {
    log.Fatal("Failed to encrypt files.", "err", err)
  }

  err = os.Remove("files.tar.gz")
  if err != nil {
    log.Fatal("Failed to delete archive.", "err", err)
  }

}

/*
 * encryptFile
 * Encrypts a file
 *
 * Takes the file to encrypt
 */
func encryptFile(file string) error {
  // Read the file and get contents
  value, err := ioutil.ReadFile(file)
  if err != nil {
    return err
  }
  
  // Create a block cipher
  // SEE: https://en.wikipedia.org/wiki/Block_cipher
  // TODO: Add actual key support
  block, err := aes.NewCipher([]byte("UTFt3YBbdTFnf5wA"))
  if err != nil {
    return err
  }
  
  // And now use GCM (Galois/Counter Mode)
  // SEE: https://en.wikipedia.org/wiki/Galois/Counter_Mode
  gcm, err := cipher.NewGCM(block)
  if err != nil {
    return err
  }
  
  // Generate a random number (nonce)
  nonce := make([]byte, gcm.NonceSize())
  if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
    return err
  }
  
  // Use the seal function (which encrypts and authenticates plaintext).
  cipherText := gcm.Seal(nonce, nonce, value, nil)
  
  // Write the file out
  err = ioutil.WriteFile(file + ".grave", cipherText, 0777)
  if err != nil {
	  return err
  }

  return nil
}


/*
 * makeArchive
 * Creates a new archive.
 *
 * Takes the name of the directory to archive.
 */
func makeArchive(dir string) error {
  // Find all the files that need to be added to the archive.
  entries, err := os.ReadDir(dir)
  if err != nil {
    return err
  }
  
  // Create a file that ends in `.tar.gz`
  buf, err := os.Create(dir + ".tar.gz")
  if err != nil {
    return err
  }

  gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

  for _, e := range entries {
    filename := dir + "/" + e.Name()

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


