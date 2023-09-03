<p align="center">
<img src="https://raw.githubusercontent.com/MilanSuk/skyalt/main/screenshots/logo.png" />
</p>

<br/><br/>
<p align="center">
<img src="https://raw.githubusercontent.com/MilanSuk/skyalt/main/screenshots/screenshot_2.png" style="border:1px solid LightGrey" />
</p>



# SkyAlt
Build local-first apps on top of .sqlite files.

https://www.skyalt.com/



# From SaaS to Local-first
Most of today's apps run in a browser as Software as a Service. Here's the list of problems you may experience:
- delay between client and server
- none or simple export
- hard to migrate between clouds
- data disappear(music playlist, etc.)
- data was tampered
- new SaaS version was released and you wanna keep using the older one
- no offline mode
- SaaS shut down
- price goes up
- 3rd party can access your data
<br/><br/>

SkyAlt solves them with [Local-first software](https://www.inkandswitch.com/local-first/). The biggest advantages can be summarized as:
- quick responses
- works offline
- ownership
- privacy(E2EE everywhere)
- works 'forever' + run any version



# From Webkit to WASM
There are few implementations of local-first platforms and most of them use Webkit. Webkit is huge and browsers are most complex things humans build and maintain. SkyAlt is heading in oposite direction - build large and complex worlds with few simple tools.

**Front-end**: Instead of writing app in HTML/CSS/JS, you pick up from many languages which compile to WASM and you use SkyAlt's apis() to draw on screen through [Immediate mode GUI](https://en.wikipedia.org/wiki/Immediate_mode_GUI) model. Every WASM file is an asset, which can be compose with other assets together into an app.

**Back-end**: There is no back-end. Front-end uses SQL to read/write data from local .sqlite files.

**Debugging**: The best tools to write and debug code are the ones developers already use. Every SkyAlt app can be compile into WASM *or* can be run as binary in separate process, which connects to SkyAlt and communicate over TCP socket. That means that developer can use any IDE and debugger, iterate quickly and compile app into wasm for final shipping.

**Formats**: For endurance, SkyAlt uses only well-known and open formats:
- WASM for binaries
- SQLite for storages
- Json for settings



## App examples
- [7Gui](https://github.com/milansuk/skyalt/blob/main/apps/7gui/main/main.go)
- [Calendar](https://github.com/milansuk/skyalt/blob/main/apps/calendar/main/main.go)
- [Map](https://github.com/milansuk/skyalt/blob/main/apps/map/main/main.go)
- [Database](https://github.com/milansuk/skyalt/blob/main/apps/db/main/main.go)



## Current state
- **Experimental**
- go-lang SDK only
- Linux / Windows / Mac(untested)
- ~13K LOC

*SkyAlt is ~40 days old. Right now, highest priority is providing best developer experience through high range of use-cases so we iterate and change apis() a lot => apps need to be edited and recompiled to wasm!*



## Compile & Run
SkyAlt is written in Go language. You can install golang from here: https://go.dev/doc/install

Dependencies(sqlite, wazero, websocket, sdl):
<pre><code>go get github.com/mattn/go-sqlite
go get github.com/tetratelabs/wazero
go get github.com/gorilla/websocket
go get github.com/veandco/go-sdl2/sdl
go get github.com/veandco/go-sdl2/ttf
go get github.com/veandco/go-sdl2/gfx
</code></pre>

SkyAlt:
<pre><code>git clone https://github.com/milansuk/skyalt
cd skyalt
go build
./skyalt
</code></pre>



## Repository
- /apps - application's repos
- /databases - user's databases
- /device - settings(dpi, language, scroll positions, etc.)
- /resources - default fonts, images for GUI



## Author
Milan Suk

Email: milan@skyalt.com

Twitter: https://twitter.com/milansuk/

*Feel free to follow or contact me with any idea, question or problem.*



## Contributing
Your feedback and code are welcome!

For bug reports or questions, please use [GitHub's issues](https://github.com/MilanSuk/skyalt/issues)

SkyAlt is licensed under **Apache v2.0** license. This repository includes 100% of the code.
