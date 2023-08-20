# Graveyard

Graveyard is inspired by [Tomb](https://dyne.org/software/tomb/) as it provides
a simple way to encrypt files. Although, with a few key differences:

 - It is written in Go, meaning it compiles to a single binary.
 - It's built on top of archives, this means there is no need for elevated permissions.
 - "Graves," are managed by Graveyard, as opposed to being untracked files.

## What was wrong with Tomb?

Now, there wasn't anything inherently wrong with Tomb. Although, I had a few
personal gripes with it, including: 

 - Being written in a scripting language it felt slow. 
 - It requires `sudo` to be installed, supposedly there was `doas` support —
   but it didn't seem to work for me.
 - It doesn't track files, this means that it could get messy (or sacrifice
   convenience) at scale.

If you feel the same way about any of these issues give Graveyard a try, but if
you don't think that any of these aren't a big deal for you then give 
[Tomb](https://dyne.org/software/tomb) a try.

## What are the goals of Graveyard? 

Graveyard has one goal: Encrypt files securely. It *does not* try to provide a 
way to send, sync, or back them up, granted all graves are stored in one
directory — so backups should be pretty simple. 

## How secure is Graveyard?

Just like Tomb, Graveyard is not a moving-parts application, so (hopefully) 
pretty secure. Although, don't take my word for it (I'm biased) read the
source code and figure out what everything is doing. 

## How do I install Graveyard?

## From GitHub Releases

Download the binary using `curl`, and then install it using `install`:

```bash
curl -LO https://github.com/BetaPictoris/graveyard/releases/latest/download/grave
sudo install -Dt /usr/local/bin -m 755 grave
```

Optionally, if you don't have root permissions, you can install it to your
user account:

```bash
install -Dt ~/.local/bin -m 755 grave
```

## From Source

You'll need a recent version of Go(lang), I tested and developed using 1.21.
To install Go on Arch Linux:

```bash
sudo pacman -Syu go
```

Or on Debian, and Ubuntu:

```bash
sudo apt install golang-go
```

Alternatively, you can download and install it directly from [Go's
website](https://go.dev/doc/install). Then to build and install:

```bash
git clone https://github.com/BetaPictoris/graveyard.git
cd graveyard
sudo make install
```

or to install to your user account:

```
make usrinstall
```

## Usage

After installation you should have the `grave` command:

```
NAME:
   grave - Dead simple encryption

USAGE:
   grave [global options] command [command options] [arguments...]

AUTHOR:
   Daniel Hall <beta@hai.haus>

COMMANDS:
   dig, new      Create a new grave
   exhume, open  Open a buried grave
   bury, close   Bury an open grave
   ls, list      List all graves
   ps, obituary  List all open graves
   help, h       Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help

```

---

[![Beta Pictoris](https://cdn.ozx.me/betapictoris/header.svg)](https://github.com/BetaPictoris)

