# tictactoe


## Problem

Create a REST based backend for a game of tic-tac-toe.

## Start the game

To run the backend, simply run

```
docker-commpose up
```


## REST API end points

* /api/v1/games (GET)- Get all games
* /api/v1/games (POST)- Start a new game
* /api/v1/games/{game_id} (GET)- Get a game
* /api/v1/games/{game_id} (PUT)- Post a new move to a game
* /api/v1/games/{game_id} (DELETE)- Delete a game

## Design decisions

* The computer randomly selects a vacant position from the board. It is a random player and does not intentionally try to win
* The state of the game is stored in a postgres sql database
* Migration scripts for the postgres database can be found in [migrations](repository/migrations) folder
* Environment variables for the game app and the db are configured in the docker-compose file and can be changed as per need
