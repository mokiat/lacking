# Lacking Game Engine

[![Go Report Card](https://goreportcard.com/badge/github.com/mokiat/lacking)](https://goreportcard.com/report/github.com/mokiat/lacking)
[![Go Reference](https://pkg.go.dev/badge/github.com/mokiat/lacking.svg)](https://pkg.go.dev/github.com/mokiat/lacking)

[![logo](logo.png)](https://mokiat.com/lacking/)

A 3D game engine / framework that lacks a lot of features, hence the name.

**WARNING** This project is still not stable. I am playing around with the code a lot and trying stuff. Avoid using it if you are looking for something serious and reliable.

As I am quickly iterating over the code and making breaking changes all the time, avoid opening Pull Requests. The best you can do, if you want to contribute, is to open an Issue. Similarly, if you plan to use it for your own project, make sure to use a stable tag and be ready to face the consequences.

## Getting Started

If you decide to give this project a try, you should give the `template` package a try. This is a quick way to set up a Hello World project.

You will need the [gonew](https://go.dev/blog/gonew) tool, which you can install as follows:

```sh
go install golang.org/x/tools/cmd/gonew@latest
```

Afterwards, you can use the following command to set up a project:

```sh
gonew github.com/mokiat/lacking/template@latest example.com/your/namespace projectdir
cd projectdir
```

You can then use `go run` to start the game:

```sh
go run ./cmd/game
```

However, the preferable way is to use the provided [Taskfile](https://taskfile.dev/). Installing the necessary CLI is easy:

```sh
go install github.com/go-task/task/v3/cmd/task@latest
```

After which, you can start the game with the following command:

```sh
task run
```

You can check for available task commands as follows:

```sh
task --list-all
```

## Documentation

A more detailed documentation (work in progress) can be found on the `lacking`'s web page: https://mokiat.com/lacking/

## Examples

I have uploaded some example games made with this engine on `itch.io`.

### Rally MKA

Drive around in a car with no particular purpose except to zone out.
Best played with keyboard or mouse. The gamepad option is hard.

[![Rally MKA](example-rally-mka.png)](https://mokiat.itch.io/rally-mka)

### AI Suppression

A Game Jam entry. Use the keyboard to defend your ship from alien airships. Users of vim will have an easy time here.

[![AI Suppression](example-ai-suppression.png)](https://mokiat.itch.io/ai-suppression)

### Dem Cows

Fly around in a plane and use a hanging club to pop cow balloons. As it uses semi-realistic aerodynamic physics it is best played with a gamepad. There wasn't enough time to balance this game so winning it with keyboard is nearly impossible.

[![Dem Cows](example-dem-cows.png)](https://mokiat.itch.io/dem-cows)
