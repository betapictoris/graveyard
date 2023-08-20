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
  "fmt"
  "path"
  
  "strings"

  "archive/tar"
  "compress/gzip"
  
  "crypto/aes"
  "crypto/rand"
  "crypto/cipher"
  "crypto/sha256"

  "github.com/urfave/cli/v2"
  "github.com/charmbracelet/log"

  "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var (
  app_path  string
  key       string
)

func main() {
  // TODO: Add an CLI option for this
  //log.SetLevel(log.DebugLevel)
  
  // Get information on where to store data
 
  // I did a Mastodon poll for this (https://mas.to/@beta/110918744607295502).
  // Use XDG_DATA_HOME/graveyard, if set.
  data_home := os.Getenv("XDG_DATA_HOME")
  if data_home == "" {
    // Otherwise, use HOME/.graveyard
    usrHome, err := os.UserHomeDir()
    if err != nil {
		  log.Fatal("Failed to get user home path.", "err", err)
    }
    
    app_path = usrHome + "/.graveyard" 
  } else {
    app_path = data_home + "/graveyard"
  }

  // Now make sure all dirs exist
  os.MkdirAll(app_path + "/graves", 0700) // This is where encrypted files will be stored.
  os.MkdirAll(app_path + "/morgue", 0700) // This is where the unencrypted files will be stored.

  // Create a placeholder text file
  if !fileExists(app_path + "/placeholder") {
    value := []byte("Hello, and welcome to your grave!\nYou can store all kinda of photos in here and they'll be secured once you close it!")
    err := os.WriteFile(app_path + "/placeholder", value, 0600)

    if err != err {
      log.Fatal("Failed to write to file", "err", err)
    }
  }


  // Create a CLI application and define commands
  app := &cli.App{
    Name:  "grave",
    Usage: "Dead simple encryption",

    Authors: []*cli.Author{
			&cli.Author{
				Name:  "Daniel Hall",
				Email: "beta@hai.haus",
			},
		},

    Commands: []*cli.Command{
      {
        Name:      "dig",
        Aliases:   []string{"new"},
        Usage:     "Create a new grave",
        ArgsUsage: "<name>",
        Action: func(cCtx *cli.Context) error {
          // Create a key from the passphrase...
          p := tea.NewProgram(initialModel())
    	    if _, err := p.Run(); err != nil {
		        log.Fatal("Failed to start Bubbletea...", "err", err)
	        }
          
          // Create a directory (this is what we'll be making the archive from)
          log.Debug("Creating directory and placeholder file...")
          os.MkdirAll(app_path + "/morgue/" + cCtx.Args().First(), 0700)
                    
          err := copyFile(app_path + "/placeholder", app_path + "/morgue/" + cCtx.Args().First() + "/readme")
          if err != nil {
            log.Fatal("Failed to copy readme.", "err", err)
          }
          
          // Create an archive from that directory.
          log.Debug("Archiving...")
          err = makeArchive(app_path + "/morgue/" + cCtx.Args().First())
          if err != nil {
            log.Fatal("Failed to archive.", "err", err)
          }

          // And encrypt it
          log.Debug("Encrypting...")
          encryptFile(app_path + "/morgue/" + cCtx.Args().First() + ".tar.gz", key)
          if err != nil {
            log.Fatal("Failed to encrypt archive.", "err", err)
          }
          
          // Now move it into the graves
          log.Debug("Moving to graveyard...")
          err = os.Rename(app_path + "/morgue/" + cCtx.Args().First() + ".tar.gz.buried", app_path + "/graves/" + cCtx.Args().First() + ".tar.gz.buried")
          if err != nil {
            log.Fatal("Failed to move to graveyard", "err", err)
          }

          // Now clean up 
          log.Debug("Cleaning...")
          os.RemoveAll(app_path + "/morgue/" + cCtx.Args().First())
          os.RemoveAll(app_path + "/morgue/" + cCtx.Args().First() + ".tar.gz")

          return nil
        },
      },
      {
        Name:      "exhume",
        Aliases:   []string{"open"},
        Usage:     "Open a buried grave",
        ArgsUsage: "<name>",
        Action: func(cCtx *cli.Context) error {
          // Create a key from the passphrase...
          p := tea.NewProgram(initialModel())
    	    if _, err := p.Run(); err != nil {
		        log.Fatal("Failed to start Bubbletea...", "err", err)
	        }

          // Decrypt the grave using that key
          log.Debug("Decrypting...")
          err := decryptFile(app_path + "/graves/" + cCtx.Args().First() + ".tar.gz.buried", key)
          if err != nil {
            log.Fatal("Failed to decrypt file.", "err", err)
          }
          
          // Move the decrypted archive into where the directory will be stored
          log.Debug("Moving to morgue...")
          err = os.Rename(app_path + "/graves/" + cCtx.Args().First() + ".tar.gz", app_path + "/morgue/" + cCtx.Args().First() + ".tar.gz")
          if err != nil {
            log.Fatal("Failed to move archive into morgue...", "err", err)
          }
          
          // Extract the archive
          log.Debug("Extracting...")
          extractArchive(app_path + "/morgue/" + cCtx.Args().First() + ".tar.gz")
          
          // Clean up
          log.Debug("Cleaning...")
          os.RemoveAll(app_path + "/morgue/" + cCtx.Args().First() + ".tar.gz")

          return nil
        },
      },
      {
        Name:      "bury",
        Aliases:   []string{"close"},
        Usage:     "Bury an open grave",
        ArgsUsage: "<name>",
        Action: func (cCtx *cli.Context) error {
          // Create a key from the passphrase...
          p := tea.NewProgram(initialModel())
    	    if _, err := p.Run(); err != nil {
		        log.Fatal("Failed to start Bubbletea...", "err", err)
	        }

          // Encrypt the grave using that key
          log.Debug("Encrypting...")
          err :=  encryptFile(app_path + "/morgue/" + cCtx.Args().First() + ".tar.gz", key)

          // Move the encrypted archive into the graves directory
          log.Debug("Moving into the graveyard...")
          err = os.Rename(app_path + "/morgue/" + cCtx.Args().First() + ".tar.gz.buried", app_path + "/graves/" + cCtx.Args().First() + ".tar.gz.buried")
          if err != nil {
            log.Fatal("Failed to move archive into morgue...", "err", err)
          }

          // Clean up
          log.Debug("Cleaning...")
          os.RemoveAll(app_path + "/morgue/" + cCtx.Args().First())

          return nil
        },
      },
    },
  }

  if err := app.Run(os.Args); err != nil {
    log.Fatal("Failed to start grave.", "err", err)
  }
}


/*
 * createKey
 * Generates a key to use for encryption
 * 
 * Takes the passphrase.
 */
func createKey(passphrase string) (key string) {
  log.Debug("Creating SHA256 checksum from passphrase...")
  // Create a checksum from the passphrase (needs to be turned into bytes)
  hash := sha256.Sum256([]byte(passphrase))
  
  // Convert the first 32 bytes into a string (this will be the key).
  key = string(hash[:32])
  
  log.Debug("Done!")
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
        log.Debug("File info dump", "path dir", path.Dir(header.Name), "name", header.Name)
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
        
        log.Debug("Copying file...", "name", header.Name)
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
  }}// 

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

// copyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFile(src, dst string) (err error) {
    in, err := os.Open(src)
    if err != nil {
        return
    }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil {
        return
    }
    defer func() {
        cerr := out.Close()
        if err == nil {
            err = cerr
        }
    }()
    if _, err = io.Copy(out, in); err != nil {
        return
    }
    err = out.Sync()
    return
}


// checks if filename exists
func fileExists(filename string) bool {
   info, err := os.Stat(filename)
   if os.IsNotExist(err) {
      return false
   }
   return !info.IsDir()
}


type model struct {
	textInput textinput.Model
	err       error
}

func initialModel() model {
	ti := textinput.New()
	//ti.Placeholder   = "Passphrase"
  ti.EchoMode      = textinput.EchoPassword
  ti.EchoCharacter = '•'
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return model{
		textInput: ti,
		err:       nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
    case tea.KeyEnter:
      key = createKey(m.textInput.Value())
      return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
      os.Exit(0)
		}

	// We handle errors just like any other message
	case error:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf(
    "Passphrase:\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}
