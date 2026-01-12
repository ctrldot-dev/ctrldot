# Installing Dot CLI

## Option 1: Add to PATH (Recommended)

Add the `bin` directory to your PATH:

```bash
# For current session
export PATH=$PATH:/Volumes/FutureBox/Code/futurematic-kernal/bin

# For permanent (add to ~/.zshrc or ~/.bashrc)
echo 'export PATH=$PATH:/Volumes/FutureBox/Code/futurematic-kernal/bin' >> ~/.zshrc
source ~/.zshrc
```

Then use:
```bash
dot --version
dot status
```

## Option 2: Install to ~/bin

```bash
# Create ~/bin if it doesn't exist
mkdir -p ~/bin

# Copy the binary
cp bin/dot ~/bin/dot
chmod +x ~/bin/dot

# Add ~/bin to PATH if not already there
echo 'export PATH=$PATH:$HOME/bin' >> ~/.zshrc
source ~/.zshrc
```

## Option 3: Install to /usr/local/bin (System-wide)

```bash
# Requires sudo
sudo cp bin/dot /usr/local/bin/dot
sudo chmod +x /usr/local/bin/dot
```

## Option 4: Create Symlink

```bash
# Create symlink in a directory that's in your PATH
ln -s $(pwd)/bin/dot ~/bin/dot

# Or in /usr/local/bin (requires sudo)
sudo ln -s $(pwd)/bin/dot /usr/local/bin/dot
```

## Verify Installation

After installation, verify it works:

```bash
dot --version
dot --help
dot status
```

## Update the Binary

When you rebuild the CLI, you'll need to update the installed version:

```bash
# If using PATH method, just rebuild
go build -o bin/dot ./cmd/dot

# If using ~/bin or /usr/local/bin, copy again
cp bin/dot ~/bin/dot
# or
sudo cp bin/dot /usr/local/bin/dot
```
