# Cleanup and Maintenance Guide

This guide covers devbox's cleanup and maintenance features to keep your development environment healthy and optimized.

## Overview

Devbox provides comprehensive cleanup and maintenance tools:

- **Cleanup**: Remove unused Docker resources and orphaned containers
- **Maintenance**: System health checks, updates, and auto-repair functionality

## Cleanup Command

The `devbox cleanup` command helps maintain a clean system by removing various Docker resources and devbox artifacts.

### Interactive Cleanup

```bash
devbox cleanup
```

This opens an interactive menu with the following options:

1. **Clean up orphaned devbox containers** - Remove containers not tracked in config
2. **Remove unused Docker images** - Remove dangling and unused images
3. **Remove unused Docker volumes** - Remove unused volumes
4. **Remove unused Docker networks** - Remove unused networks
5. **Run Docker system prune** - Comprehensive cleanup of all unused resources
6. **Clean up everything** - Combines options 1-4
7. **Show system status** - Display disk usage and system information

### Command-Line Flags

```bash
# Specific cleanup tasks
devbox cleanup --orphaned           # Remove orphaned containers only
devbox cleanup --images             # Remove unused images only
devbox cleanup --volumes            # Remove unused volumes only
devbox cleanup --networks           # Remove unused networks only
devbox cleanup --system-prune       # Run docker system prune
devbox cleanup --all                # Clean up everything

# Safety and information
devbox cleanup --dry-run            # Show what would be cleaned (no changes)
devbox cleanup --force              # Skip confirmation prompts
```

### Examples

```bash
# See what would be cleaned without making changes
devbox cleanup --dry-run --all

# Clean only orphaned containers
devbox cleanup --orphaned

# Comprehensive cleanup with confirmation
devbox cleanup --all

# Quick cleanup without prompts
devbox cleanup --all --force
```

## Maintenance Command

The `devbox maintenance` command provides system health monitoring, updates, and repair functionality.

### Interactive Maintenance

```bash
devbox maintenance
```

This opens an interactive menu with these options:

1. **Check system status** - Show Docker status, projects, and disk usage
2. **Perform health check** - Check health of all projects
3. **Update system packages** - Update packages in all containers
4. **Restart stopped containers** - Start any stopped devbox containers
5. **Rebuild all containers** - Recreate containers from latest base images
6. **Auto-repair common issues** - Automatically fix detected problems
7. **Full maintenance** - Combines health check, updates, and restarts

### Command-Line Flags

```bash
# Individual maintenance tasks
devbox maintenance --status         # Show detailed system status
devbox maintenance --health-check   # Check health of all projects
devbox maintenance --update         # Update all containers
devbox maintenance --restart        # Restart stopped containers
devbox maintenance --rebuild        # Rebuild all containers
devbox maintenance --auto-repair    # Auto-fix common issues

# Control flags
devbox maintenance --force          # Skip confirmation prompts
```

### Examples

```bash
# Check system health
devbox maintenance --health-check

# Update all containers
devbox maintenance --update

# Rebuild all containers (with confirmation)
devbox maintenance --rebuild

# Quick full maintenance without prompts
devbox maintenance --force --health-check --update --restart
```

## Health Checks

The health check system monitors:

- **Container Status**: Whether containers are running or stopped
- **Container Responsiveness**: Whether containers respond to commands
- **Workspace Directories**: Whether project directories exist
- **Configuration Files**: Whether devbox.json files are valid

Health check results show:
- ✅ **Healthy**: Container running and responsive
- ⚠️ **Unhealthy**: Container stopped or unresponsive
- ❌ **Missing**: Container or workspace missing

## Auto-Repair

The auto-repair feature automatically fixes common issues:

- **Missing workspace directories**: Creates missing project directories
- **Missing containers**: Recreates containers from configuration
- **Stopped containers**: Starts stopped containers
- **Unresponsive containers**: Restarts containers that don't respond

## System Updates

The update feature:
1. Runs `apt update -y` to refresh package lists
2. Runs `apt full-upgrade -y` to install updates
3. Runs `apt autoremove -y` to remove unnecessary packages
4. Runs `apt autoclean` to clean package cache

Updates are applied to all tracked containers that are running or can be started.

## Container Rebuilding

The rebuild feature:
1. Stops and removes existing containers
2. Pulls latest base images
3. Recreates containers with current configuration
4. Runs system updates
5. Executes setup commands from devbox.json
6. Sets up devbox environment

**Note**: Rebuilding preserves your project files but recreates the container environment.

## Best Practices

### Regular Maintenance

Run these commands regularly:

```bash
# Weekly: Health check and updates
devbox maintenance --health-check --update

# Monthly: Full cleanup
devbox cleanup --all

# As needed: Auto-repair issues
devbox maintenance --auto-repair
```

### Monitoring Disk Usage

```bash
# Check Docker disk usage
devbox cleanup
# Select option 7 to show system status

# Or use maintenance status
devbox maintenance --status
```

### Before Major Changes

```bash
# Before system upgrades or major changes
devbox maintenance --health-check
devbox cleanup --dry-run --all

# After making changes
devbox maintenance --auto-repair
```

## Troubleshooting

### Common Issues and Solutions

**Orphaned Containers**
```bash
# Problem: Containers exist but aren't tracked
devbox cleanup --orphaned

# Or remove specific container
docker rm -f devbox_oldproject
```

**Disk Space Issues**
```bash
# Comprehensive cleanup
devbox cleanup --all

# System prune for maximum cleanup
devbox cleanup --system-prune
```

**Container Won't Start**
```bash
# Check what's wrong
devbox maintenance --health-check

# Try auto-repair
devbox maintenance --auto-repair

# Manual rebuild if needed
devbox maintenance --rebuild
```

**Configuration Problems**
```bash
# Check project configuration
devbox config validate myproject

# Show current configuration
devbox config show myproject
```

### Safe Cleanup

Always use `--dry-run` first to see what would be removed:

```bash
# Safe: See what would be cleaned
devbox cleanup --dry-run --all

# Then run actual cleanup
devbox cleanup --all
```

### Emergency Recovery

If something goes wrong:

```bash
# Check system status
devbox maintenance --status

# Try auto-repair
devbox maintenance --auto-repair

# Rebuild problematic projects
devbox destroy myproject
devbox init myproject
```

## Integration with Docker

These commands work alongside standard Docker commands:

```bash
# Devbox cleanup
devbox cleanup --images

# Equivalent Docker command
docker image prune -f

# Devbox system status
devbox maintenance --status

# Equivalent Docker commands
docker system df
docker ps -a
```

The devbox cleanup and maintenance commands are designed to be safe wrappers around Docker operations, with additional intelligence about devbox project structure and configuration.