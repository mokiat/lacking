---
title: Getting Started
---

# Getting Started

The easiest way to set up a new project is to use the `template` package. Beforehand, make sure you have the necessary prerequisites.

* Ensure the [Go glfw bindings](https://github.com/go-gl/glfw?tab=readme-ov-file#installation) run on your platform.

* Ensure you have the [gonew](https://go.dev/blog/gonew) tool.

    ```
    go install golang.org/x/tools/cmd/gonew@latest
    ```

* Ensure you have the [task](https://taskfile.dev/) tool.

    ```
    go install github.com/go-task/task/v3/cmd/task@latest
    ```

Afterwards, you can use the following steps:

1. Create new project using the Lacking template.

    ```
    gonew github.com/mokiat/lacking-template@latest example.com/your/namespace projectdir
    ```

    ```
    cd projectdir
    ```

1. Build and package the assets.

    ```
    task pack
    ```

1. Run the project.

    ```
    task run
    ```

You would only need to run `task pack` initially and whenever the source resources (images, models) or the pipeline that transforms them have been changed.

You can check for available task commands as follows:

```sh
task --list-all
```
