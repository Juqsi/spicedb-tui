# spicedb-tui

**spicedb-tui** is a modern terminal user interface (TUI) for [SpiceDB](https://spicedb.dev), offering an alternative to [zed](https://github.com/authzed/zed) for interactive management and exploration—directly from your terminal.
It provides fast, efficient, and multilingual access to all core features of SpiceDB.

You can get started quickly by [downloading pre-built binaries](https://github.com/Juqsi/spicedb-tui/releases) for your platform, or by building the tool yourself.

---

Falls du es lieber noch kürzer, noch sachlicher, oder auf Deutsch möchtest, sag Bescheid!
---

## Features

* View, edit, and upload schema definitions
* Browse, filter, create, and delete relationships (tuples)
* Batch and filtered deletes with regex support
* Check permissions interactively
* Create and restore SpiceDB backups
* Multilingual interface (currently English & German, easily extendable)
* Configurable connection and language from within the TUI

---

## Installation

1. **Clone the repository**

   ```sh
   git clone https://github.com/juqsi/spicedb-tui.git
   cd spicedb-tui
   ```

2. **Build**

   ```sh
   go build -o spicedb-tui ./cmd
   ```

3. **Run**

   ```sh
   ./spicedb-tui
   ```

   Or for development:

   ```sh
   go run ./cmd
   ```

---

## Configuration

On first start, a `config.json` will be created in the working directory.
Configure your SpiceDB endpoint, token, and preferred language either through the TUI or by editing the file directly.

---

## Requirements

* Go 1.20 or newer
* A running [SpiceDB](https://spicedb.dev) instance
* A valid preshared token

---

## License

MIT License

---

## Credits

* [SpiceDB](https://spicedb.dev) – Authorization at scale
* [rivo/tview](https://github.com/rivo/tview) – Powerful terminal UI framework for Go

---

Feedback, issues, and contributions are welcome.
