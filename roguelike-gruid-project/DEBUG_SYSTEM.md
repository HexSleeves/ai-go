# Debug Visualization System

This document describes the comprehensive debug visualization system implemented for the roguelike game. The system provides visual debugging tools for FOV (Field of View), AI pathfinding, and AI decision-making processes.

## Features

### 1. FOV Debug Visualization
- **Currently visible tiles**: Bright white/yellow color
- **Previously explored tiles**: Dim gray color  
- **Completely unexplored tiles**: Very dark/black color
- Overrides normal map colors when enabled
- Maintains wall/floor distinction with debug color scheme

### 2. AI Pathfinding Debug Visualization
- **Color-coded paths based on AI behavior**:
  - **Red paths**: Aggressive/chase behavior (pursuing the player)
  - **Yellow paths**: Flee/retreat behavior (moving away from player)
  - **Blue paths**: Patrol/wander behavior
  - **Green paths**: Searching/other behaviors
- **Target destination markers**: Diamond symbols (◆) in magenta
- **Failed pathfinding attempts**: X markers (×) in red
- **Path visualization**: Uses Unicode line drawing characters (─, │, ·)

### 3. AI Decision Tree Visualization
- **Side panel display** showing AI entity information
- **Real-time AI state tracking** (Chasing, Fleeing, Patrolling, etc.)
- **Key decision factors**:
  - Distance to player
  - Health percentage
  - Line of sight status
  - Last known player position
  - Search turns remaining
- **Sorted by proximity** to player (closest entities first)
- **Limited display** (entities within 20 tiles, max panel size)

## Key Bindings

| Key | Function | Description |
|-----|----------|-------------|
| **F1** | Toggle Pathfinding Debug | Shows/hides pathfinding paths and targets |
| **F2** | Print Pathfinding Stats | Outputs pathfinding statistics to console |
| **F3** | Toggle FOV Debug | Shows/hides FOV state visualization |
| **F4** | Toggle AI Debug | Shows/hides AI decision panel |
| **F5** | Cycle Debug Levels | Cycles through combined debug modes |

## Debug Levels

The system supports 5 debug levels that can be cycled with F5:

1. **None (0)**: All debug visualizations disabled
2. **FOV Only (1)**: Only FOV debug visualization
3. **Pathfinding Only (2)**: Only pathfinding debug visualization  
4. **AI Only (3)**: Only AI decision debug panel
5. **Full Debug (4)**: All debug visualizations enabled

## Color Scheme

### FOV Debug Colors
- `ColorDebugFOVVisible`: Bright white (currently visible)
- `ColorDebugFOVExplored`: Dim gray (previously explored)
- `ColorDebugFOVUnexplored`: Very dark (unexplored)

### Pathfinding Debug Colors
- `ColorDebugPathChasing`: Red (aggressive behavior)
- `ColorDebugPathFleeing`: Yellow (retreat behavior)
- `ColorDebugPathPatrolling`: Blue (patrol/wander behavior)
- `ColorDebugPathSearching`: Green (searching behavior)
- `ColorDebugTarget`: Magenta (target destinations)
- `ColorDebugWaypoint`: Cyan (waypoints)

### AI Debug Panel Colors
- `ColorDebugAIPanel`: Gray (panel text and borders)
- `ColorUIText`: Default (entity headers)
- `ColorUITitle`: Bright (panel title)

## Implementation Details

### Architecture
- **Modular design**: Each debug mode can be enabled independently
- **Performance optimized**: Debug info only collected when modes are active
- **Viewport aware**: Only renders debug info for visible areas
- **Memory efficient**: Uses cached debug information updated per turn

### Key Components

#### Debug System (`debug.go`)
- `DebugLevel` enum for debug modes
- `AIDebugInfo` structure for AI debug data
- `AIEntityDebug` structure for individual entity debug info
- Utility functions for colors, strings, and path characters

#### Model Integration (`model.go`)
- Debug state management
- Toggle methods for each debug mode
- Debug level cycling
- Debug info updates per turn

#### Rendering Integration (`rendering.go`)
- FOV debug overlay in map rendering
- Enhanced pathfinding debug with AI behavior colors
- AI debug panel rendering
- Efficient screen coordinate conversion

#### Pathfinding Enhancement (`pathfinding.go`)
- Extended `PathfindingDebugInfo` with AI state information
- AI behavior tracking for color coding
- Integration with existing pathfinding debug system

## Usage Examples

### Basic Debug Usage
```go
// Toggle individual debug modes
model.ToggleFOVDebug()        // F3
model.ToggleAIDebug()         // F4
model.TogglePathfindingDebug() // F1

// Cycle through debug levels
model.CycleDebugLevel()       // F5
```

### Accessing Debug Information
```go
// Get current debug info
debugInfo := model.GetDebugInfo()
aiDebugInfo := model.GetAIDebugInfo()
pathfindingDebugInfo := model.GetPathfindingDebugInfo()
```

### Color Utilities
```go
// Get appropriate debug colors
fovColor := GetFOVDebugColor(isVisible, isExplored)
pathColor := GetPathfindingDebugColor(aiState)

// Get human-readable strings
stateString := GetAIStateString(aiState)
behaviorString := GetAIBehaviorString(aiBehavior)
```

## Performance Considerations

- **Conditional rendering**: Debug overlays only drawn when enabled
- **Efficient updates**: Debug info cached and updated only when needed
- **Viewport culling**: Only processes entities and tiles in camera view
- **Memory management**: Debug structures cleaned up when disabled
- **Minimal overhead**: No performance impact when debug modes are off

## Release Build Considerations

The debug system is designed to be easily removable for release builds:

- All debug code is contained in specific files (`debug.go`, debug methods)
- Debug rendering is conditionally executed
- Debug key handlers can be disabled with build tags
- No debug overhead in normal gameplay when disabled

## Troubleshooting

### Common Issues
1. **Debug panel not visible**: Check screen size - panel requires at least 25 characters width
2. **No pathfinding paths shown**: Ensure entities have both AI and pathfinding components
3. **FOV debug not working**: Verify player has FOV component
4. **Performance issues**: Disable debug modes when not needed

### Debug Information
Use F2 to print pathfinding statistics to console for detailed debugging information.

## Future Enhancements

Potential improvements to the debug system:
- Configurable debug colors
- Debug info export/logging
- More detailed AI decision factors
- Performance profiling integration
- Debug command console
- Save/load debug configurations
