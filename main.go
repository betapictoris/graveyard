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
  "runtime"
  "errors" 
  "strings"

  "archive/tar"
  "compress/gzip"

  "encoding/base64"
  
  "crypto/aes"
  "crypto/rand"
  "crypto/cipher"
  "crypto/subtle"
  
  "golang.org/x/crypto/argon2"

  "github.com/urfave/cli/v2"
  "github.com/charmbracelet/log"

  "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var (
  app_path      string

  key           string
  current_grave string
  grave_is_new  bool
)

func main() {
  // TODO: Add an CLI option for this
  log.SetLevel(log.WarnLevel)
  
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

    Flags: []cli.Flag{
      &cli.StringFlag{
        Name:  "log",
        Value: "warn",
        Usage: "Logging level",
      },
    },
    Before: func(cCtx *cli.Context) error {
      if cCtx.String("log") != "warn" {
        log.SetLevel(log.DebugLevel)
      }
      return nil
   },

    Commands: []*cli.Command{
      {
        Name:      "dig",
        Aliases:   []string{"new"},
        Usage:     "Create a new grave",
        ArgsUsage: "<name>",
        Action: func(cCtx *cli.Context) error {
          current_grave = cCtx.Args().First() // TODO: Refactor other uses of this param.
          grave_is_new  = true

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
          err = encryptFile(app_path + "/morgue/" + cCtx.Args().First() + ".tar.gz", key)
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
          current_grave = cCtx.Args().First() // TODO: Refactor other uses of this param.

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

          fmt.Println(app_path + "/morgue/" + cCtx.Args().First())

          return nil
        },
      },
      {
        Name:      "bury",
        Aliases:   []string{"close"},
        Usage:     "Bury an open grave",
        ArgsUsage: "<name>",
        Action: func (cCtx *cli.Context) error {
          current_grave = cCtx.Args().First() // TODO: Refactor other uses of this param.

          // Create a key from the passphrase...
          p := tea.NewProgram(initialModel())
    	    if _, err := p.Run(); err != nil {
		        log.Fatal("Failed to start Bubbletea...", "err", err)
	        }

          // Archive the directory
          log.Debug("Archiving...")
          err := makeArchive(app_path + "/morgue/" + cCtx.Args().First())

          // Encrypt the grave using that key
          log.Debug("Encrypting...")
          err = encryptFile(app_path + "/morgue/" + cCtx.Args().First() + ".tar.gz", key)

          // Move the encrypted archive into the graves directory
          log.Debug("Moving into the graveyard...")
          err = os.Rename(app_path + "/morgue/" + cCtx.Args().First() + ".tar.gz.buried", app_path + "/graves/" + cCtx.Args().First() + ".tar.gz.buried")
          if err != nil {
            log.Fatal("Failed to move archive into morgue...", "err", err)
          }

          // Clean up
          log.Debug("Cleaning...")
          os.RemoveAll(app_path + "/morgue/" + cCtx.Args().First())
          os.RemoveAll(app_path + "/morgue/" + cCtx.Args().First() + ".tar.gz")

          return nil
        },
      },
      {
        Name:       "ls",
        Aliases:    []string{"list"},
        Usage:      "List all graves",
        Action: func (cCtx *cli.Context) error {
          // Find all graves
          log.Debug("Searching for graves...")
          entries, _ := os.ReadDir(app_path + "/graves/")

          // Now print each
          for _, e := range entries {
            // We are using the logger here so it looks cleaner, and we don't
            // have to worry about the logging level. 
            fmt.Println(strings.ReplaceAll(e.Name(), ".tar.gz.buried", "")) 
          }
          
          return nil
        },
      },
      {
        Name:     "ps",
        Aliases:  []string{"obituary"},
        Usage:    "List all open graves",
        Action: func (cCtx *cli.Context) error {
          // Find all open graves
          log.Debug("Searching for open graves...")
          entries, _ := os.ReadDir(app_path + "/morgue/")

          // Now print each
          for _, e := range entries {
            // See "ls"/"list" on why we're using fmt
            fmt.Println(e.Name())
          }

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
 * checkKey
 * Checks if a passphrase is correct.
 *
 * Takes a grave and passphrase.
 * Returns a key and error.
 */
func checkKey(grave, passphrase string) (string, error) {
  log.Info("Checking key...")

  // Get all keys from the keys file. :
  log.Debug("Reading keys file...")
  keys, err := ioutil.ReadFile(app_path + "/keys")
  if err != nil {
    return "", err
  }

  var encoded_hash string
  
  log.Debug("Finding key...")
  // Loop through all keys -- this isn't perfect, but it works.
  for _, i := range strings.Split(string(keys), "\n") {
    e := strings.Split(i, " ")

    g := e[0] // The grave will be the first section of the line
    h := e[1] // and the hash is on the second section

    if g == grave {
      encoded_hash = h
    }
  }
  
  // Try to parse the encoded hash... 
  log.Debug("Getting values...")
  vals := strings.Split(encoded_hash, "$")
  if len(vals) != 6 {
    return "", errors.New("The hash is not in the correct format.")
  }
  
  var version int
  log.Debug("Checking version...")
  _, err = fmt.Sscanf(vals[2], "v=%d", &version)
  if err != nil {
    return "", err
  }
  if version != argon2.Version {
    return "", errors.New("The argon2 version is incompatible.")
  }
  
  var memory int
  var iterations int
  var threads int
  log.Debug("Finding memory, iterations, and threads...")
  _, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &memory, &iterations, &threads)
  if err != nil {
    return "", err
  }
  
  log.Debug("Finding salt...")
  salt, err := base64.RawStdEncoding.Strict().DecodeString(vals[4])
  if err != nil {
    return "", err
  }
  //saltLength := uint32(len(salt))
  
  log.Debug("Finding hash...")
  hash, err := base64.RawStdEncoding.Strict().DecodeString(vals[5])
  if err != nil {
    return "", err
  }
  keyLength := uint32(len(hash))
  
  log.Debug("Rehashing with the same parameters...")
  newHash := argon2.IDKey([]byte(passphrase), salt, uint32(iterations), uint32(memory), uint8(threads), keyLength)
  
  log.Debug("Checking...")
  if subtle.ConstantTimeCompare(hash, newHash) == 1 {
    return string(newHash), nil
  }
  return "", errors.New("Authentication denied")
}

/* 
 * createKey
 * Generates a key to use for encryption
 * 
 * Takes the grave and passphrase.
 * Returns a key and error.
 */
func createKey(grave, passphrase string) (string, error) {
  log.Debug("Creating key from passphrase...")
  // Now we will build an Argon2 hash on top of this
  
  // To do this we first need to make a salt.
  log.Debug("Creating salt...")
  salt := make([]byte, 16)
  _, err := rand.Read(salt)
  if err != nil {
    return "", err
  }
  
  // Find the number of cores (as the number of threads while hashing)
  threads := runtime.NumCPU() 

  // Use the argon IDkey function, with double the RFC recommendations.
  log.Debug("Creating argon2 hash...")
  argon := argon2.IDKey([]byte(passphrase), salt, (1 * 2), ((64 * 1024) * 2), uint8(threads), 32)

  // Encode for saving...
  log.Debug("Encoding salt...")
  b64Salt := base64.RawStdEncoding.EncodeToString(salt)
  log.Debug("Encoding hash...")
  b64Hash := base64.RawStdEncoding.EncodeToString(argon)

  // Save to the keys file
  log.Debug("Parsing hash...")
  encodedHash := fmt.Sprintf("%s $argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", grave, argon2.Version, ((64 * 1024) * 2), (1 * 2), threads, b64Salt, b64Hash)
  
  log.Debug("Opening keys file...")
  f, err := os.OpenFile(app_path + "/keys", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
  if err != nil {
	  return "", err
  }
  defer f.Close()

  log.Debug("Writing encoded hash...")
  if _, err := f.WriteString(encodedHash); err != nil {
	  return "", err
  }

  log.Debug("Done!")
  return string(argon), nil
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
      // This is where we'll check the user's input
      // (on the enter key press)
      
      var err error

      if grave_is_new {
        key, err = createKey(current_grave, m.textInput.Value())
      } else {
        key, err = checkKey(current_grave, m.textInput.Value())
      }

      if err != nil {
        log.Fatal("Failed to preform key action.", "is new key", grave_is_new, "err", err)
      }
      
      if key == "" {
        log.Fatal("Couldn't validate key.")
      }

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
