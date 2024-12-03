# Audio to MIDI Converter

A FastAPI-based application to convert audio files from a specified path into MIDI files.

## Features

- Convert audio files to MIDI format.
- FastAPI server for handling conversion requests.
- Simple and lightweight, with no database integration.

## Prerequisites

- Python 3.11.4 or higher

## Setup

1. Create a virtual environment:

   ```bash
   python -m venv .venv
   ```

2. Activate the virtual environment:

   - On Windows (git bash):

     ```bash
     source .venv/Scripts/activate
     ```

3. Install the required dependencies:
   ```bash
   pip install -r requirements.txt
   ```

## Usage (Development)

1. Start the FastAPI development server:

   ```bash
   fastapi dev main.py
   ```
