# spicui

**spicui** is a spicy terminal user interface (TUI) for [SpiceDB](https://spicedb.dev).  
Manage and explore your SpiceDB instance directly from the terminal – modern, colorful, efficient, and optionally multilingual.

---

## Features

-  **View and upload schema**
-  **Browse, create, and delete relationships (tuples)**
-  **Check permissions interactively**
-  **Create and restore backups**
-  **Import demo or example data**
-  **Multilingual interface (e.g., English & German)**
-  **Async operations with loading indicators (the TUI stays responsive!)**
-  **Chili-inspired look and easy keyboard navigation**

---

## Installation

1. **Clone the project**
    ```sh
    git clone https://github.com/juqsi/spicui.git
    cd spicui
    ```

2. **Build**
    ```sh
    go build -o spicui ./cmd
    ```

3. **Run**
    ```sh
    ./spicui
    ```
   Or for development:
    ```sh
    go run ./cmd
    ```

---

## Configuration

On first start, a `config.json` will be created in the project directory.  
You can set your SpiceDB endpoint, token, and language – or edit these settings directly in the TUI.

---

## Requirements

- Go 1.20 or newer
- A running [SpiceDB](https://spicedb.dev) instance local
- A valid preshared token

---

## License

MIT License  
Built with 🌶️ and love.

---

## Credits

- [SpiceDB](https://spicedb.dev) – the best authorization backend
- [rivo/tview](https://github.com/rivo/tview) – for an awesome TUI framework

---

*Questions, suggestions, PRs or issues? All welcome!*
