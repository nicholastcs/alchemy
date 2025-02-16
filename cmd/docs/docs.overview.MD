## Overview

Alchemy is a flexible platform-agnostic, golden pattern code generator.

## Concepts

### Form
A declarative code to define the form used by consumer (Software developers). The form is powered by:
- [Charmbracelet Huh Terminal Form](https://github.com/charmbracelet/huh).

### FormResult 
A internal API to define the results from the Form, after a Form is filled.

### CodeTemplate 
A declarative code for Code Template, which it facilitates to generate files and folders 
on host after consuming FormResult. 

For now it will support for `go-template`.
