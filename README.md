# GoLoad - Modern Download Manager

GoLoad is a powerful, lightweight download manager built with Go and React. It offers a modern user interface with real-time updates and efficient download management.

## Features

- **Fast Downloads**: Optimized for performance with efficient file handling
- **Concurrent Downloads**: Download multiple files simultaneously
- **Pause/Resume**: Ability to pause and resume downloads at any time
- **Speed Control**: Limit download speeds globally or per download
- **Modern UI**: Clean, responsive interface with dark mode support
- **Real-time Updates**: Live progress tracking and statistics
- **Cross-platform**: Works on Windows, macOS, and Linux
- **Desktop & Web Versions**: Use as a desktop app or web service

## Installation

### Desktop App

1. Download the latest release for your platform from the [Releases](https://github.com/TOomaAh/GoLoad/releases) page
2. Run the installer and follow the on-screen instructions
3. Launch GoLoad from your applications menu

### From Source

#### Prerequisites
- Go 1.18+
- Node.js 16+
- Wails CLI (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

#### Build Steps
```bash
# Clone the repository
git clone https://github.com/TOomaAh/GoLoad.git
cd GoLoad

# Install frontend dependencies
cd frontend
npm install
cd ..

# Build the application
wails build
```

## Usage

### Adding Downloads

1. Click the "New" button in the top-right corner
2. Enter the URL of the file you want to download
3. Click "Download" to start

### Managing Downloads

- **Pause**: Click the pause button to temporarily stop a download
- **Resume**: Click the play button to resume a paused download
- **Cancel**: Click the X button to cancel an ongoing download
- **Remove**: Click the trash button to remove a download from the list

### Settings

Access the settings menu to configure:
- Default download location
- Maximum concurrent downloads
- Global speed limit
- Auto-start behavior
- Retry options

## Development

### Project Structure
```
GoLoad/
├── internal/                # Backend Go code
│   ├── core/                # Core application logic
│   │   ├── download/        # Download management
│   │   └── settings/        # Application settings
│   └── utils/               # Utility functions
├── frontend/                # React frontend
│   ├── src/                 # Source code
│   │   ├── components/      # React components
│   │   └── DownloadManager.jsx # Main component
│   └── package.json         # Frontend dependencies
└── wails.json               # Wails configuration
```

### Running in Development Mode

```bash
# Start the development server
wails dev
```

## Technical Details

GoLoad uses:
- Go for the backend with GORM for database operations
- React with Tailwind CSS for the frontend
- Wails for desktop integration
- SQLite for persistent storage

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [Wails](https://wails.io/) for the Go/JavaScript runtime
- [GORM](https://gorm.io/) for the ORM library
- [React](https://reactjs.org/) for the frontend framework
- [Tailwind CSS](https://tailwindcss.com/) for styling
- [Lucide React](https://lucide.dev/) for the icons

---

Built with ❤️ by TOomaAh