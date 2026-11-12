# grassrook

> **Warning:** Grassrook is currently under active development and is still in its early stages. Features may be incomplete, unstable, or subject to significant changes.

An all-in-one desktop application for managing drifting competitions, from tournament organization and driver registration to judging and battle management.

## Build from source

### Prerequisites

- [Go](https://go.dev/dl/)
- [Node.js](https://nodejs.org/)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation/)
- Platform-specific Wails dependencies

### Development

Clone the repository:

```bash
git clone https://github.com/grassrook/client grassrook
cd grassrook
```

Install frontend dependencies:

```bash
cd frontend
npm install
cd ..
```

Run the application in development mode:

```bash
wails dev
```

### Production build

Build the application for your current platform:

```bash
wails build
```

The compiled application will be available in the `build/bin` directory.
