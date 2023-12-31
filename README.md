# spongebob-cli

SpongeBob delivered straight from your terminal

![example](https://github.com/overflowy/spongebob-cli/assets/98480250/c3280a20-bc16-40c5-b1d8-6ce4602331f1)

## Why?

Why not.

## Features

- List all available episodes
- Stream episodes directly with minimal user interaction
- Customize the video player used for streaming
- Download all episodes asynchronously

## Usage

Running spongebob-cli without any flags will prompt the user to select the episode number.

```
Usage of spongebob-cli:
  -af int
        add an episode to your favourites
  -d int
        download all episodes asynchronously but max [d] episodes at a time (default -1)
  -l    list episodes and quit
  -lf
        list favourite episodes
  -p int
        play the wanted episode without any user interaction (default -1)
  -rf int
        remove an episode from your favourites
  -vp string
        use another video player [default=mpv] (default "mpv")
```

## Disclaimer

This tool is for educational purposes only. The maintainers do not own the rights to any of the content streamed by this application. It is the user's responsibility to ensure they have the right to watch the streamed content.
