## Server

If you have the application running, the next step is to configure the backend.

Just as the frontend which is light and cross platform, the backend is actually an image.  
It makes it more easier to deploy, manage and update.

The backend receives data from Powens API and store them in a mySQL database.

## MySQL database
First, you have to create a mySQL database.  

Instructions are [located here](https://dev.mysql.com/doc/mysql-getting-started/en/){:target="_blank"}.

When your mySQL database is up and running, you can initialize the table.  
For now, there are 6 of them.  

You can copy / paste the command [from the migration folder](https://github.com/soragXYZ/freenahi/tree/main/backend/migrations){:target="_blank"} directly in a console, or source the files if you cloned the repo.

```shell
source /<yourPath>/freenahi/backend/migrations/authToken.sql
source /<yourPath>/freenahi/backend/migrations/bankAccount.sql
source /<yourPath>/freenahi/backend/migrations/historyValue.sql
source /<yourPath>/freenahi/backend/migrations/invest.sql
source /<yourPath>/freenahi/backend/migrations/loan.sql
source /<yourPath>/freenahi/backend/migrations/tx.sql
```


## Container image

To set up the backend, you need to have an engine container installed on your system (possibly different that the one which hosts the application).  

You can install Podman. [More information here](https://podman.io/docs/installation){:target="_blank"} if needed.
!!! quote ""
    Podman is a daemonless, open source, Linux native tool designed to make it easy to find, run, build, share and deploy applications using Open Containers Initiative (OCI) Containers and Container Images.

???+ tip
    Docker can be used too, but Podman is known to be a better alternative than Docker

## The image
The image is hosted [here on the docker hub](https://hub.docker.com/r/soragxyz/freenahi/tags){:target="_blank"}

To download the image:



=== "podman"

    ```shell
    podman pull soragxyz/freenahi:X.Y.Z
    ```

=== "docker"

    ```shell
    docker pull soragxyz/freenahi:X.Y.Z
    ```

Before running the image, you need to configure multiple environment variables :

Name                   | Role                                 | Example value |
---                    | ---                                  | --- |
POWENS_CLIENT_ID       | Credentials to connect to Powens API | XXXXX |
POWENS_CLIENT_SECRET   | Credentials to connect to Powens API | XXXXX |
POWENS_DOMAIN          | Credentials to connect to Powens API | XXXXX |
POWENS_WEBVIEW_URL     | The URL webview                      | https://webview.powens.com/ |
POWENS_REDIRECT_URL    | The redirection link for the webview | https://xxxx/ |
POWENS_WHITELISTED_IPS | The whitelisted IPs for your backend | 127.0.0.1,::1,13.39.29.243,15.188.68.198,13.39.95.239 |
DB_NAME                | Your DDB name                        | XXXXX |
DB_HOST                | Your DDB IP                          | localhost |
DB_PORT                | Your DDB port                        | 3306 |
DB_USER                | Your DDB username                    | XXXXX |
DB_PASS                | Your DDB password                    | XXXXX |
SERVER_PORT            | The port used by the server          | 8080 |
SERVER_TIMEOUT_READ    | Server config for timeout            | 3s |
SERVER_TIMEOUT_WRITE   | Server config for timeout            | 5s |
SERVER_TIMEOUT_IDLE    | Server config for timeout            | 5s |
SERVER_LOG_LEVEL       | The logs level                       | trace |
OTHER_LANGUAGE         | The langage for the webview          | en |


If needed, there is an environment example file [located here](https://github.com/soragXYZ/freenahi/blob/main/backend/.env.exemple){:target="_blank"}.
You need to update your environment variables according to your configuration.  

You can get powens environment variables when creating your account. See [Powens page](./powens.md)

???+ danger
    For now, there is no authentication mecanism, which means that if an IP is in the whitelisted IPs, it has full access and can query the backend to retrieve data.  
    It is recommended to run the application and the server locally at first.

## Start the backend 

Now that you have configured your Powens account, the database and the server, you should be able to start the server

To start it, you have set the environment variables.
You can manually specify a file where every environment variable is declared :  

=== "podman"

    ```shell
    podman run --env-file <pathToEnv> soragxyz/freenahi:X.Y.Z
    ```

=== "docker"

    ```shell
    docker run --env-file <pathToEnv> soragxyz/freenahi:X.Y.Z
    ```

???+ tip
    If you don't want to use a file, you can specify every environment variable with [the option --env](https://docs.podman.io/en/v5.0.1/markdown/podman-run.1.html#env-e-env){:target="_blank"}.  

    Also, the option **--network=host** might be usefull if you are running the application and the server on the same machine.