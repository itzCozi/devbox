---
title: Introduction
description: Learn about devbox and its principles
---

Welcome to the devbox documentation! This guide provides an overview of what devbox is, its core principles, and how it can help you manage your development environments effectively. devbox is a tool designed to create isolated, reproducible development environments using containers. It simplifies the process of setting up and managing dependencies, ensuring that your projects run consistently across different machines.

#### Key Features of devbox
- **Isolation**: Each project gets its own isolated environment, preventing dependency conflicts.
- **Reproducibility**: Environments can be easily recreated on any machine, ensuring consistent behavior.
- **Simplicity**: Easy to set up and manage environments with simple commands.
- **Flexibility**: Supports a wide range of programming languages and frameworks.
- **Portability**: Environments can be shared and versioned alongside your code.

## The Problem
---

Traditional development environments can lead to "it works on my machine" issues, where code behaves differently depending on the local setup. This can cause significant delays and frustration, especially in team settings where multiple developers work on the same project on different machines with different dependencies and configurations.

## Principles of devbox
---

1. **Containerization**: Each development environment runs in a container, ensuring isolation from the host system and other projects.
2. **Configuration as Code**: Environments are defined using configuration files, allowing for version control and easy sharing.
3. **On-Demand Environments**: Environments are created and destroyed as needed, reducing resource usage.
4. **User-Centric Design**: Focused on developer experience, making it easy to switch between projects and manage dependencies.
5. **Scope In**: Focus on simplicity and core functionality, avoiding unnecessary complexity in favor of stability.

## Getting Started

To get started with devbox, follow this guide on [installation and setup](/start).
