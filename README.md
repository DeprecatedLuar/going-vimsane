# going-vimsane

Visual X11 overlay that shows colored borders when switching between kanata keyboard layers. Works with the included `vimsanity.kbd` config for vim-like keyboard navigation.

## Requirements

- X11 (not Wayland)
- kanata keyboard remapper
- Go 1.25+ (for building)
- X11 dev libraries: `libx11-dev`, `libxext-dev`, `libxrender-dev`

## Quick Setup

### 1. Install kanata

Download the latest release:
```bash
wget https://github.com/jtroo/kanata/releases/latest/download/kanata -O /usr/local/bin/kanata
chmod +x /usr/local/bin/kanata
```

Or build from source:
```bash
git clone https://github.com/jtroo/kanata
cd kanata
cargo build --release
sudo cp target/release/kanata /usr/local/bin/
```

### 2. Set up permissions (no sudo required)

```bash
# Create uinput group and add your user
sudo groupadd uinput
sudo usermod -aG input $USER
sudo usermod -aG uinput $USER

# Add udev rule
echo 'KERNEL=="uinput", MODE="0660", GROUP="uinput", OPTIONS+="static_node=uinput"' | \
  sudo tee /etc/udev/rules.d/99-input.rules

# Reload and apply
sudo udevadm control --reload-rules && sudo udevadm trigger
sudo modprobe uinput

# Log out and back in for group changes to take effect
```

### 3. Set up kanata service

**Quick one-liner setup:**
```bash
mkdir -p ~/.config/kanata && cp vimsanity.kbd ~/.config/kanata/ && sudo tee /etc/systemd/system/kanata.service > /dev/null <<'EOF'
[Unit]
Description=Kanata keyboard remapper
Documentation=https://github.com/jtroo/kanata

[Service]
Type=simple
ExecStart=/usr/local/bin/kanata --cfg %h/.config/kanata/vimsanity.kbd --port 5829
Restart=on-failure
RestartSec=3

[Install]
WantedBy=default.target
EOF
sudo systemctl daemon-reload && sudo systemctl enable kanata.service && sudo systemctl start kanata.service
```

**Or manually:**

Copy `vimsanity.kbd` to your config directory:
```bash
mkdir -p ~/.config/kanata
cp vimsanity.kbd ~/.config/kanata/
```

Create `/etc/systemd/system/kanata.service`:
```ini
[Unit]
Description=Kanata keyboard remapper
Documentation=https://github.com/jtroo/kanata

[Service]
Type=simple
ExecStart=/usr/local/bin/kanata --cfg %h/.config/kanata/vimsanity.kbd --port 5829
Restart=on-failure
RestartSec=3

[Install]
WantedBy=default.target
```

**Note**: `%h` expands to the user's home directory in systemd

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable kanata.service
sudo systemctl start kanata.service
```

### 4. Build the overlay

```bash
go build -o going-vimsane
```

### 5. Set up overlay service

Edit `going-vimsane.service` and update the path, then:
```bash
# Copy to systemd user directory
mkdir -p ~/.config/systemd/user/
cp going-vimsane.service ~/.config/systemd/user/

# Enable and start
systemctl --user daemon-reload
systemctl --user enable going-vimsane.service
systemctl --user start going-vimsane.service
```

## How it works

- **kanata** runs with `--port 5829` to expose layer changes via TCP
- **going-vimsane** connects to kanata's TCP server and displays colored borders based on the active layer
- Each vim layer (normal, visual, delete, etc.) gets a unique color
- Border disappears when on the `default` layer

## Layer Colors

- `vim-normal` → Green
- `visual-mode` → Magenta
- `vim-shifted` → Red
- `delete-ops` → Red
- `yank-ops` → Yellow
- `g-ops` → Cyan
- `meta-layer` → Purple
- `escape` → Gray
- `default` → Hidden

## Customization

Edit constants in `overlay.go`:
- `TintAlpha` - Screen tint opacity (0-255)
- `BorderWidth` - Border thickness in pixels
- `CornerRadius` - Corner rounding radius

## Troubleshooting

**Overlay not showing:**
- Check kanata is running: `systemctl status kanata`
- Verify TCP port: `ss -tlnp | grep 5829`
- Check overlay logs: `systemctl --user status going-vimsane`

**Kanata permission denied:**
- Ensure you're in `input` and `uinput` groups: `groups`
- Log out and back in after adding groups
- Check udev rule: `cat /etc/udev/rules.d/99-input.rules`

**Build errors:**
- Install X11 dev packages: `sudo apt install libx11-dev libxext-dev libxrender-dev`
