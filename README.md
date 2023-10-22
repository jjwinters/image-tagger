<img src="ImageViewer/data/icon.png" width=64 height=64> 

# Image Tagger 

## About

This is a simple image viewer lets you quickly iterate through a directory and tag (rename) image files using configurable buttons.

Commissioned for a friend for this specific purpose, hence the default tags. 

It also includes some basic editing functionality carried over from Palexer/image-viewer. 

Graciously forked from [image-viewer](https://github.com/Palexer/image-viewer)


## Screenshot

![Screenshot](screenshot.png)

## Used Tools

- language: Go
- UI framework: fyne
- image processing backend: gift
- cross compilation: fyne-cross

## Installation

The executable is currently available for Windows. Just download the archive from the [releases section](https://github.com/jjwinters/image-tagger/releases), extract and install it as usual.

Compiling other OS/architectures are possible if desired, I just haven't tried it and my testing options are limited.

## Contributing

1. Fork this repository to your local machine
2. Add a new branch, make changes, merge to master branch
3. Make a pull request

## Development

### Linux (Ubuntu)

The following packages are required:
- gcc
- pkg-config
- libgl1-mesa-dev
- xorg-dev
- [fyne-cross](https://github.com/fyne-io/fyne-cross)    (for cross compile)
- docker        (for fyne-cross)

## ToDo

- Hotkeys for tag buttons.

## Bugs

- Windows may give Permission Denied error when attempting to rename files. This may have to do with file permissions or antivirus.

## Help

If you need any help, feel free to open a new [Issue](https://github.com/jjwinters/image-tagger/issues).

## License

[MIT](LICENSE)
