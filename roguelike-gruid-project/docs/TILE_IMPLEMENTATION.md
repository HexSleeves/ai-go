# Tile-Based Rendering Implementation

This document describes the implementation of tile-based rendering in the roguelike game, providing an upgrade path from ASCII characters to graphical tiles.

## Overview

The tile system provides a seamless upgrade from ASCII to graphical rendering while maintaining full backward compatibility. Players can switch between ASCII and tile modes at runtime, and the system gracefully handles missing tiles with automatic fallbacks.

## Architecture

### Core Components

1. **TileConfig** (`internal/config/tiles.go`)
   - Configuration management for tile settings
   - Persistent storage of user preferences
   - Runtime configuration updates

2. **TileMapping** (`internal/ui/tile_mapping.go`)
   - Maps game entities (runes) to tile image files
   - Supports multiple tiles per entity for variety
   - Handles special cases for different entity types

3. **ImageTileManager** (`internal/ui/image_tiles.go`)
   - Implements gruid.TileManager interface
   - Manages tile loading, caching, and color processing
   - Provides fallback to font-based rendering

4. **Fallback System** (`internal/ui/fallback_tiles.go`)
   - Generates basic tile images when tileset is missing
   - Ensures the game always works even without external assets

## Key Features

### 1. Dual Rendering Support
- **ASCII Mode**: Traditional character-based rendering using fonts
- **Tile Mode**: Image-based rendering using PNG tiles
- **Runtime Switching**: Toggle between modes with 'T' key
- **Automatic Fallback**: Falls back to ASCII if tiles fail to load

### 2. Flexible Tile Mapping
```go
// Multiple tiles per entity for visual variety
tm.EntityTiles['o'] = []string{"monsters/orc.png", "monsters/orc_variant.png"}

// Special handling for different entity types
func (tm *TileMapping) GetTileForEntity(entityID ecs.EntityID, ecsSystem *ecs.ECS) string
```

### 3. Performance Optimization
- **Tile Caching**: LRU cache with configurable size limits
- **Color Caching**: Pre-processed color variants
- **Lazy Loading**: Tiles loaded on-demand
- **Memory Management**: Automatic cache eviction

### 4. Configuration System
```json
{
  "tiles": {
    "tiles_enabled": false,
    "tile_size": 16,
    "scale_factor": 1.0,
    "tileset_path": "assets/tiles/",
    "use_smoothing": true,
    "cache_size": 1000
  }
}
```

## Implementation Details

### Build System Integration

The tile system uses Go build tags to provide different implementations:

```go
//go:build js || sdl
// +build js sdl
```

- **SDL/JS builds**: Full tile support with ImageTileManager
- **Default builds**: ASCII-only with stub functions

### Entity-to-Tile Mapping

The mapping system supports multiple approaches:

1. **Direct Rune Mapping**: `'@'` → `"characters/player.png"`
2. **Entity-Specific**: Based on entity components and state
3. **Fallback Chain**: Rune → Entity Type → Generic → ASCII

### Color Processing

Tiles are processed to apply game colors:

```go
func (itm *ImageTileManager) applyColors(baseImg image.Image, fg, bg gruid.Color, attrs gruid.AttrMask) image.Image
```

- Foreground/background color application
- Style attribute handling (reverse, etc.)
- Transparency support

### Memory Management

The tile cache implements several optimization strategies:

- **Size Limits**: Configurable maximum cache size
- **LRU Eviction**: Removes least recently used tiles
- **Preloading**: Common tiles loaded at startup
- **Color Variants**: Cached separately to avoid reprocessing

## Usage

### For Players

1. **Enable Tiles**: Use `--tiles` flag or press 'T' in-game
2. **Download Tileset**: Follow instructions in `assets/tiles/README.md`
3. **Configure**: Modify settings in `~/.config/roguelike-gruid/tile_config.json`

### For Developers

1. **Add New Entities**: Update tile mapping in `tile_mapping.go`
2. **Custom Tiles**: Set `TileName` field in Renderable component
3. **Performance Tuning**: Adjust cache size and preloading strategy

## Recommended Tileset

**Kenney's Roguelike/RPG Pack**
- 1,700+ tiles, 16×16 pixels
- CC0 License (Public Domain)
- Complete coverage of roguelike elements
- Download: https://kenney.nl/assets/roguelike-rpg-pack

## File Structure

```
internal/
├── config/
│   └── tiles.go              # Tile configuration
├── ui/
│   ├── image_tiles.go        # Main tile manager (SDL/JS)
│   ├── tile_mapping.go       # Entity-to-tile mapping
│   ├── fallback_tiles.go     # Fallback tile generation
│   ├── tiles_stub.go         # Stubs for non-SDL builds
│   └── sdl.go               # SDL driver integration
└── game/
    ├── player.go            # Toggle tile action handler
    └── input.go             # Key binding for 'T'

assets/tiles/
├── characters/              # Player and NPC tiles
├── monsters/               # Monster tiles
├── environment/            # Walls, floors, doors
├── items/                  # Weapons, potions, etc.
├── ui/                     # UI elements
├── fallback/               # Auto-generated fallbacks
└── README.md               # Setup instructions
```

## Error Handling

The system provides robust error handling:

1. **Missing Tiles**: Falls back to ASCII rendering
2. **Invalid Configuration**: Uses sensible defaults
3. **Memory Pressure**: Automatic cache eviction
4. **File System Errors**: Graceful degradation

## Performance Considerations

### Memory Usage
- Typical cache: 50-100MB for 1000 tiles
- Color variants: Additional 2-3x memory per tile
- Configurable limits prevent excessive usage

### Loading Performance
- Initial load: 100-500ms for common tiles
- Runtime loading: <10ms per tile (cached)
- Preloading reduces in-game hitches

### Rendering Performance
- Tile rendering: Comparable to font rendering
- Color processing: Optimized with caching
- Scale factor: Linear impact on memory/performance

## Future Enhancements

1. **Animated Tiles**: Support for multi-frame animations
2. **Tile Atlases**: Single image with multiple tiles
3. **Dynamic Loading**: Stream tiles from network
4. **Compression**: Reduce memory usage with compressed formats
5. **GPU Acceleration**: Hardware-accelerated rendering

## Troubleshooting

### Common Issues

1. **Tiles not showing**: Check build tags and SDL2 installation
2. **Performance issues**: Reduce cache size or scale factor
3. **Memory usage**: Enable cache size limits
4. **Missing tiles**: Check tileset directory structure

### Debug Information

Enable debug logging to see tile system activity:
```bash
./roguelike --debug
```

This will show:
- Tile loading attempts and failures
- Cache hit/miss statistics
- Memory usage information
- Configuration loading status
