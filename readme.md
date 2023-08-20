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
 - It didn't track files, this means that it could get messy (or sacrifice
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
