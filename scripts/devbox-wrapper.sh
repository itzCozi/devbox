#!/bin/bash




BOX_NAME="${DEVBOX_BOX_NAME:-unknown}"
PROJECT_NAME="${DEVBOX_PROJECT_NAME:-unknown}"

case "$1" in
    "exit"|"quit")
        echo "Exiting devbox shell for project '$PROJECT_NAME'"
        exit 0
        ;;
    "status"|"info")
    echo "Devbox box status"
        echo "Project: $PROJECT_NAME"
        echo "Box: $BOX_NAME"
        echo "Workspace: /workspace"
        echo "Host: $(cat /etc/hostname)"
        echo "User: $(whoami)"
        echo "Working Directory: $(pwd)"
        echo ""
    echo "hint: available devbox commands inside box:"
        echo "  devbox exit     - Exit the shell"
        echo "  devbox status   - Show box information"
        echo "  devbox help     - Show this help"
        echo "  devbox host     - Run command on host (experimental)"
        ;;
    "help"|"--help"|"-h")
        echo "Devbox box commands"
        echo ""
        echo "Available commands inside the box:"
        echo "  devbox exit         - Exit the devbox shell"
        echo "  devbox status       - Show box and project information"
        echo "  devbox help         - Show this help message"
        echo "  devbox host <cmd>   - Execute command on host (experimental)"
        echo ""
    echo "Your project files are in: /workspace"
    echo "You're in an Ubuntu box with full package management"
        echo ""
        echo "Examples:"
        echo "  devbox exit                    # Exit to host"
        echo "  devbox status                  # Check box info"
        echo "  devbox host 'devbox list'     # Run host command"
        echo ""
    echo "hint: Files in /workspace are shared with your host system"
        ;;
    "host")
        if [ -z "$2" ]; then
            echo "error: usage: devbox host <command>"
            echo "Example: devbox host 'devbox list'"
            exit 1
        fi
        echo "Executing on host: $2"
        echo "warning: This is experimental and may not work in all environments"


    echo "error: host command execution not yet implemented"
    echo "hint: Exit the box and run commands on the host instead"
        ;;
    "version")
        echo "devbox box wrapper v1.0"
        echo "Box: $BOX_NAME"
        echo "Project: $PROJECT_NAME"
        ;;
    "")
        echo "error: missing command. Use 'devbox help' for available commands."
        exit 1
        ;;
    *)
        echo "error: unknown devbox command: $1"
        echo "hint: Use 'devbox help' to see available commands inside the box"
        echo ""
        echo "Available commands:"
        echo "  exit, status, help, host, version"
        exit 1
        ;;
esac
