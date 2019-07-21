# Multiplayer Snake via the Terminal

*_Work in Progress!_*

```sh
git clone https://github.com/iamlouk/gosnake.git
cd gosnake
go build

# Server:
./gosnake --addr "localhost:1234" --server

# Player 1:
./gosnake --addr "localhost:1234" --nick player1 --peer player2

# Player 2:
./gosnake --addr "localhost:1234" --nick player2 --peer player1

```
