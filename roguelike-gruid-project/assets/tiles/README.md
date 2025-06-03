# Tileset Setup

This directory contains the tile assets for the roguelike game. The game supports both ASCII and tile-based rendering.

## Recommended Tileset: Kenney's Roguelike/RPG Pack

The game is designed to work with **Kenney's Roguelike/RPG Pack**, which provides 1,700+ high-quality 16x16 pixel tiles.

### Download Instructions

1. **Visit the download page**: https://kenney.nl/assets/roguelike-rpg-pack
2. **Download the pack** (it's completely free under CC0 license)
3. **Extract the contents** to this directory

### Expected Directory Structure

After downloading and extracting, your directory structure should look like:

```
assets/tiles/
├── characters/
│   ├── knight_m.png
│   ├── knight_f.png
│   └── ... (other character tiles)
├── monsters/
│   ├── orc.png
│   ├── goblin.png
│   ├── skeleton.png
│   └── ... (other monster tiles)
├── environment/
│   ├── wall_mid.png
│   ├── floor_1.png
│   ├── door_closed.png
│   └── ... (other environment tiles)
├── items/
│   ├── flask_red.png
│   ├── scroll_01.png
│   ├── weapon_sword.png
│   └── ... (other item tiles)
├── ui/
│   └── ... (UI elements)
└── fallback/
    └── ... (automatically generated fallback tiles)
```

### Alternative Tilesets

While the game is optimized for Kenney's tileset, you can use other 16x16 tilesets by:

1. Organizing tiles in the same directory structure
2. Updating the tile mapping in `internal/ui/tile_mapping.go`
3. Ensuring tile names match the expected filenames

### Fallback System

If tiles are missing, the game will:

1. **Generate basic fallback tiles** automatically in the `fallback/` directory
2. **Fall back to ASCII rendering** if tile loading fails
3. **Continue working** without any crashes

### Configuration

- **Enable/Disable tiles**: Press `T` in-game or use the `--tiles` command-line flag
- **Tile configuration**: Stored in `~/.config/roguelike-gruid/tile_config.json`
- **Runtime switching**: Toggle between ASCII and tiles with the `T` key

### Building with Tile Support

```bash
# Build with SDL2 support (required for tiles)
go build -tags sdl

# Build traditional ASCII version
go build
```

### Troubleshooting

**Tiles not showing?**
- Ensure you've built with `-tags sdl`
- Check that SDL2 is installed on your system
- Verify tiles are in the correct directory structure
- Check the console output for tile loading errors

**Performance issues?**
- Reduce tile cache size in configuration
- Use smaller scale factor
- Disable tile smoothing

**Missing tiles?**
- The game will generate basic fallback tiles automatically
- Check `fallback/` directory for generated tiles
- Missing tiles will fall back to ASCII characters

### License

Kenney's Roguelike/RPG Pack is released under **CC0 (Public Domain)** license, meaning you can use it for any purpose without attribution (though attribution is appreciated).

### Support

For tileset-related issues:
- Check the game's debug output (`--debug` flag)
- Verify your directory structure matches the expected layout
- Ensure you have proper file permissions
- Try regenerating fallback tiles by deleting the `fallback/` directory
