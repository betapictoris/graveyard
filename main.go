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
  "path"
  
  "strings"

  "archive/tar"
  "compress/gzip"
  
  "crypto/aes"
  "crypto/rand"
  "crypto/cipher"
  "crypto/sha256"
  
  "github.com/charmbracelet/log"
)

func main() {
  key := createKey("pass")

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

  err = os.RemoveAll("files.tar.gz.buried")
  if err != nil {
    log.Fatal("Failed to remove file.", "err", err)
  }

  err = os.RemoveAll("files")
  if err != nil {
    log.Fatal("Failed to remove directory.", "err", err)
  }

  err = extractArchive("files.tar.gz")
  if err != nil {
    log.Fatal("Failed to extract archive.", "err", err)
  }

  err = os.Remove("files.tar.gz")
  if err != nil {
    log.Fatal("Failed to remove file.", "err", err)
  }
}


/*
 * createKey
 * Generates a key to use for encryption
 * 
 * Takes the passphrase.
 */
func createKey(passphrase string) (key string) {
  hash := sha256.Sum256([]byte(passphrase))
  key = string(hash[:32]) // Create a 32-bit key

  return string(key)
}

/*
 * extractAcrhive
 * Decompresses and unarchive a .tar.gz
 * 
 * Takes the file's path the unarchive.
 */
func extractArchive(file string) error {
  // First open the file to obtain a reader
  log.Debug("Opening file...", "file", file)
  reader, err := os.Open(file)
  
  // Open a gunzip (gz) reader
  log.Debug("Opening gz reader...")
  gr, err := gzip.NewReader(reader)
  if err != nil {
    return err
  }

  // And now a tar reader
  log.Debug("Opening tar reader...")
  tr := tar.NewReader(gr)

  // Start an infinite loop on the file
  for true {
    // Grab the header 
    log.Debug("Reading content's header info...")
    header, err := tr.Next()
    

    if err == io.EOF {
      // If we have reached the end of the file (EOF) then we can safely break
      // the loop.
      log.Debug("EOF, Done!")
      break
    } else if err != nil {
      // Otherwise, if the error is not nil, return the error.
      return err
    }

    // New we'll create a switch statement to handle the different types of archive
    // contents.
    switch header.Typeflag {
      case tar.TypeDir:
        log.Debug("Creating directory...", "name", header.Name)
        // If the header is a directory recreate it.
        if err := os.Mkdir(header.Name, 0755); err != nil {
          return err
        }
      case tar.TypeReg:
        log.Debug("Creating directories, to ensure they exists...", "path", path.Dir(header.Name))
        // Create directory path to ensure it exists... 
        os.MkdirAll(path.Dir(header.Name), 0700)
        
        log.Debug("Creating file...", "name", header.Name)
        // If it's a file, create a file with the same name.
        file, err := os.Create(header.Name)
        defer file.Close() // Close the file when we're done.
        if err != nil {
          return err
        }
        
        log.Debug("Copying file...", "namelemmy.blahaj.zone", header.Name)
        // Copy over the file.
        if _, err := io.Copy(file, tr); err != nil {
          return err
        }
      default:
        // If it's not a directory or a file then we're in a strange place where
        // we don't know what to do. 
        // We could handle this in two ways, returning an error or warning the
        // user, in this case we'll warn the user (as the other files may still
        // be known types).
        log.Warn("File is of unknown type.", "type", header.Typeflag)
  }}

  log.Info("The body has been removed from the casket!", "stored at", strings.ReplaceAll(file, ".tar.gz", ""))
  return nil
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

 
