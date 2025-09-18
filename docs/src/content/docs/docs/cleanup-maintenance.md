---
title: Cleanup and Maintenance
description: Keep your devbox development environment healthy and optimized
---

This guide covers devbox's cleanup and maintenance features to keep your development environment healthy and optimized.

## Cleanup Command
---

The `devbox cleanup` command helps maintain a clean system by removing various Docker resources and devbox artifacts.

##### Interactive Cleanup

```bash
devbox cleanup
```

This opens an interactive menu with the following options:

1. **Clean up orphaned devbox boxes** - Remove boxes not tracked in config
2. **Remove unused Docker images** - Remove dangling and unused images
3. **Remove unused Docker volumes** - Remove unused volumes
4. **Remove unused Docker networks** - Remove unused networks
5. **Run Docker system prune** - Comprehensive cleanup of all unused resources
6. **Clean up everything** - Combines options 1-4
7. **Show system status** - Display disk usage and system information

##### Command-Line Flags

```bash
# Specific cleanup tasks
devbox cleanup --orphaned           # Remove orphaned boxes only
devbox cleanup --images             # Remove unused images only
devbox cleanup --volumes            # Remove unused volumes only
devbox cleanup --networks           # Remove unused networks only
devbox cleanup --system-prune       # Run docker system prune
devbox cleanup --all                # Clean up everything

# Safety and information
devbox cleanup --dry-run            # Show what would be cleaned (no changes)
devbox cleanup --force              # Skip confirmation prompts
```

##### Examples

```bash
# See what would be cleaned without making changes
devbox cleanup --dry-run --all

# Clean only orphaned boxes
devbox cleanup --orphaned

# Comprehensive cleanup with confirmation
devbox cleanup --all

# Quick cleanup without prompts
devbox cleanup --all --force
```

## Maintenance Command
---

The `devbox maintenance` command provides system health monitoring, updates, and repair functionality.

##### Interactive Maintenance

```bash
devbox maintenance
```

This opens an interactive menu with these options:

1. **Check system status** - Show Docker status, projects, and disk usage
2. **Perform health check** - Check health of all projects
3. **Update system packages** - Update packages in all boxes
4. **Restart stopped boxes** - Start any stopped devbox boxes
5. **Rebuild all boxes** - Recreate boxes from latest base images
6. **Auto-repair common issues** - Automatically fix detected problems
7. **Full maintenance** - Combines health check, updates, and restarts

##### Command-Line Flags

```bash
# Individual maintenance tasks
devbox maintenance --status         # Show detailed system status
devbox maintenance --health-check   # Check health of all projects
devbox maintenance --update         # Update all boxes
devbox maintenance --restart        # Restart stopped boxes
devbox maintenance --rebuild        # Rebuild all boxes
devbox maintenance --auto-repair    # Auto-fix common issues

# Control flags
devbox maintenance --force          # Skip confirmation prompts
```

##### Examples

```bash
# Check system health
devbox maintenance --health-check

# Update all boxes
devbox maintenance --update

# Rebuild all boxes (with confirmation)
devbox maintenance --rebuild

# Quick full maintenance without prompts
devbox maintenance --force --health-check --update --restart
```

## Update Command
---

Use the `devbox update` command to rebuild environment boxes from the latest base images. This is the recommended way to apply upstream image updates or configuration changes that affect the base image or setup commands.

##### Why use `devbox update`?

- Pulls the newest base image(s)
- Recreates the box with your current `devbox.json` configuration
- Automatically runs a full system update inside the box
- Re-runs your `setup_commands` to ensure tools are present
- Preserves your project files on the host at `~/devbox/<project>/`

##### Usage

```bash
# Update a single project
devbox update myproject

# Update all projects
devbox update
```

##### When to use maintenance vs update

- `devbox maintenance --update`: Update system packages inside existing boxes
- `devbox update`: Rebuild boxes from the latest base images and re-apply configuration

If you're changing `base_image` in `devbox.json` or want to ensure you are using the latest upstream image, use `devbox update`.

## Health Checks
---

The health check system monitors:

- **Box Status**: Whether boxes are running or stopped
- **Box Responsiveness**: Whether boxes respond to commands
- **Workspace Directories**: Whether project directories exist
- **Configuration Files**: Whether devbox.json files are valid

Health check results show:
- ✅ **Healthy**: Box running and responsive
- ⚠️ **Unhealthy**: Box stopped or unresponsive
- ❌ **Missing**: Box or workspace missing

## Auto-Repair
---

The auto-repair feature automatically fixes common issues:

- **Missing workspace directories**: Creates missing project directories
- **Missing boxes**: Recreates boxes from configuration
- **Stopped boxes**: Starts stopped boxes
- **Unresponsive boxes**: Restarts boxes that don't respond

## System Updates
---

The update feature:
1. Runs `apt update -y` to refresh package lists
2. Runs `apt full-upgrade -y` to install updates
3. Runs `apt autoremove -y` to remove unnecessary packages
4. Runs `apt autoclean` to clean package cache

Updates are applied to all tracked boxes that are running or can be started.

## Box Rebuilding
---

The rebuild feature:
1. Stops and removes existing boxes
2. Pulls latest base images
3. Recreates boxes with current configuration
4. Runs system updates
5. Executes setup commands from devbox.json
6. Sets up devbox environment

:::caution
Rebuilding preserves your project files but recreates the box environment.
:::

## Monitoring
---

##### System Status
```bash
# Quick status overview
devbox list

# Detailed system information
devbox maintenance --status

# Docker resource usage
docker system df
```

##### Box Health
```bash
# Check all project health
devbox maintenance --health-check

# Check specific box
docker inspect devbox_myproject

# View box logs
docker logs devbox_myproject
```

##### Resource Usage
```bash
# Live box stats
docker stats

# Disk usage by type
docker system df -v

# List all devbox boxes
docker ps -a --filter "name=devbox_"
```
