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
  
  "strings"

  "archive/tar"
  "compress/gzip"
  
  "crypto/aes"
  "crypto/rand"
  "crypto/cipher"
  
  //tea "github.com/charmbracelet/bubbletea"
  "github.com/charmbracelet/log"
)

func main() {
  // TODO: Add actual key support
  key := "BT4BGmmr3ZCKeTw8"

  err := makeArchive("files")
  if err != nil {
    log.Fatal("Failed to create archive.", "err", err)
  }

  err = encryptFile("files.tar.gz", key)
  if err != nil {
    log.Fatal("Failed to encrypt files.", "err", err)
  }

  err = os.Remove("files.tar.gz")
  if err != nil {
    log.Fatal("Failed to delete archive.", "err", err)
  }
  
  err = decryptFile("files.tar.gz.buried", key)
  if err != nil {
    log.Fatal("Failed to decrypt archive.", "err", err)
  }

}

/*
 * decryptFile
 * Decrypts a file
 *
 * Takes the file to decrypt
 */
func decryptFile(file, key string) error {
  // Read the file and get the contents
  log.Debug("Reading file for decryption...")
  value, err := ioutil.ReadFile(file)
  if err != nil {
    return err
  }

  // Recreate the block cipher
  log.Debug("Recreating block cipher...")
  block, err := aes.NewCipher([]byte(key))
  if err != nil {
    return err
  }

  // Reset up the GCM
  log.Debug("Using GCM mode...")
  gcm, err := cipher.NewGCM(block)
  if err != nil {
    return err
  }
  
  // Now we'll need to get the nonce we created
  log.Debug("Finding nonce...")
  nonce := value[:gcm.NonceSize()]

  // After we have the nonce we can get the actual value and open the file
  log.Debug("Getting unencrypted value...")
  value = value[gcm.NonceSize():]
  plainValue, err := gcm.Open(nil, nonce, value, nil)
  if err != nil {
    return err
  }

  // Finally, write out the file
  log.Debug("Writing file...")
  err = ioutil.WriteFile(strings.ReplaceAll(file, ".buried", ""), plainValue, 0777)

  log.Info("The casket has been exhumed!", "exhumed at", strings.ReplaceAll(file, ".buried", ""))
  return err
}

/*
 * encryptFile
 * Encrypts a file
 *
 * Takes the file to encrypt
 */
func encryptFile(file, key string) error {
  // Read the file and get contents
  log.Debug("Reading file for encryption...", "file", file)
  value, err := ioutil.ReadFile(file)
  if err != nil {
    return err
  }
  
  // Create a block cipher
  // SEE: https://en.wikipedia.org/wiki/Block_cipher
  log.Debug("Creating cipher from key...")
  block, err := aes.NewCipher([]byte(key))
  if err != nil {
    return err
  }
  
  // And now use GCM (Galois/Counter Mode)
  // SEE: https://en.wikipedia.org/wiki/Galois/Counter_Mode
  log.Debug("Using GCM mode...")
  gcm, err := cipher.NewGCM(block)
  if err != nil {
    return err
  }
  
  // Generate a random number (nonce)
  log.Debug("Generating nonce...")
  nonce := make([]byte, gcm.NonceSize())
  if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
    return err
  }
  
  // Use the seal function (which encrypts and authenticates plaintext).
  log.Debug("Encrypting file...")
  cipherText := gcm.Seal(nonce, nonce, value, nil)
  
  // Write the file out
  log.Debug("Writing file...")
  err = ioutil.WriteFile(file + ".buried", cipherText, 0777)
  if err != nil {
	  return err
  }

  log.Info("The file has been buried!", "buried at", file + ".buried")

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
  log.Debug("Reading files...")
  entries, err := os.ReadDir(dir)
  if err != nil {
    return err
  }
  
  // Create a file that ends in `.tar.gz`
  log.Debug("Creating files...", "file", dir + ".tar.gz")
  buf, err := os.Create(dir + ".tar.gz")
  if err != nil {
    return err
  }
  
  log.Debug("Opening the new file...")
  gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

  for _, e := range entries {
    filename := dir + "/" + e.Name()
    log.Debug("Adding file...", "file", filename)

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
  
  log.Info("The body is in the casket!", "casket", dir + ".tar.gz")
  return nil
}


