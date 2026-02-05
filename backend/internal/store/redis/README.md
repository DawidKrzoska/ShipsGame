# Redis Game State Schema

Key prefix: `game:{id}`

## Key Map

- `game:{id}:meta` (HASH)
- `game:{id}:board:p1` (HASH)
- `game:{id}:board:p2` (HASH)
- `game:{id}:occupancy:p1` (HASH)
- `game:{id}:occupancy:p2` (HASH)
- `game:{id}:ships:p1` (HASH)
- `game:{id}:ships:p2` (HASH)
- `game:{id}:shots:p1` (HASH)
- `game:{id}:shots:p2` (HASH)
- `game:{id}:events` (LIST, optional)
- `game:join:{joinCode}` (STRING -> gameId)

## Meta Hash Fields

- `id` string
- `join_code` string
- `status` = `waiting|placing|active|finished`
- `turn` = `p1|p2`
- `winner` = `p1|p2|""`
- `p1_ready` = `0|1`
- `p2_ready` = `0|1`
- `p1_joined` = `0|1`
- `p2_joined` = `0|1`
- `p1_remaining` = total ship cells remaining (int)
- `p2_remaining` = total ship cells remaining (int)

## Board Hashes

Ships per player are stored in hashes as JSON-encoded coordinate arrays.

Field format: `ship:{type}`
Value: `[[row,col],[row,col],...]`

Example:
- `ship:destroyer` -> `[[0,0],[0,1]]`

## Shots Hashes

Shots per player are stored as hashes keyed by `"row,col"`.

Value format:
- `miss`
- `hit`
- `sunk:{type}`

Example:
- `"3,5"` -> `hit`
- `"7,2"` -> `sunk:submarine`

## Occupancy Hashes

Stores per-cell ship ownership for fast lookup during `fire` updates.

Field format: `"row,col"`
Value: ship type string (e.g., `destroyer`)

## Ships Hashes

Stores remaining health per ship for sunk detection.

Field format: ship type string (e.g., `destroyer`)
Value: remaining cells (int)

## Notes

- Updates to ship placement and firing should be atomic via Lua scripts or `WATCH/MULTI`.
- Use `game:join:{joinCode}` to resolve a join code into a `gameId`.
