<h1>Extending and Customizing Environments</h1>
ddev provides several ways in which the environment for a project using ddev can be customized and extended.

## Changing PHP version

The project's `.ddev/config.yaml` file defines the PHP version to use. This can be changed, and the php_version can be set there to (currently) "5.6", "7.0", "7.1" or "7.2". The current default is php 7.1.

## Adding services to a project

For most standard web applications, ddev provides everything you need to successfully provision and develop a web application on your local machine out of the box. More complex and sophisticated web applications, however, often require integration with services beyond the standard requirements of a web and database server. Examples of these additional services are Apache Solr, Redis, Varnish, etc. While ddev likely won't ever provide all of these additional services out of the box, it is designed to provide simple ways for the environment to be customized and extended to meet the needs of your project.

A collection of vetted service configurations is available in the [Additional Services Documentation](additional-services.md).

If you need to create a service configuration for your project, see [Defining an additional service with Docker Compose](custom-compose-files.md)

## Providing custom environment variables to a container

Each project can have an unlimited number of .ddev/docker-compose.*.yaml files as described in [Custom Compose Files](./custom-compose-files.md), so it's easy to maintain custom environment variables in a .ddev/docker-compose.environment.yaml file (the exact name doesn't matter, if it just matches docker-compose.*.yaml).

For example, a `.ddev/docker-compose.environment.yaml` with these contents would add a $TYPO3_CONTEXT environment variable to the web container, and a $SOMETHING environment variable to the db container: 

```
version: '3'

services:
  web:
    environment:
      - TYPO3_CONTEXT=Development
  db:
    environment:
      - SOMETHING=something special
```

## Providing custom nginx configuration
The default web container for ddev uses NGINX as the web server. A default configuration is provided in the web container that should work for most Drupal 7+ and WordPress projects. Some projects may require custom configuration, for example to support a module or plugin requiring special rules. To accommodate these needs, ddev provides a way to replace the default configuration with a custom version.

- Run `ddev config` for the project if it has not been used with ddev before.
- Create a file named "nginx-site.conf" in the ".ddev" directory for your project. In the ddev web container, these are separate per project type, and you can see and copy the default configurations in the [web container code](https://github.com/drud/docker.nginx-php-fpm-local/tree/master/files/etc/nginx). You can also use `ddev ssh` to review existing configurations in the container at /etc/nginx.
- Add your configuration changes to the "nginx-site.conf" file. You can optionally use the [ddev nginx configs](https://github.com/drud/docker.nginx-php-fpm-local/tree/master/files/etc/nginx) as a starting point. [Additional configuration examples](https://www.nginx.com/resources/wiki/start/#other-examples) and documentation are available at the [nginx wiki](https://www.nginx.com/resources/wiki/)
- **NOTE:** The "root" statement in the server block must be `root $NGINX_DOCROOT;` in order to ensure the path for NGINX to serve the project from is correct.
- Save your configuration file and run `dev start` or `ddev restart` to start the project. If you encounter issues with your configuration or the project fails to start, use `ddev logs` to inspect the logs for possible NGINX configuration errors (or use `ddev ssh` and inspect /var/log/nginx/error.log.)
- Any errors in your configuration may cause the web container to fail and try to restart, so if you see that behavior, use `ddev logs` to diagnose.
- **IMPORTANT**: Changes to .ddev/nginx-site.conf take place on a `ddev restart`.

## Providing custom PHP configuration (php.ini)

You can provide additional PHP configuration for a project by creating a directory called `.ddev/php/` and adding any number of php configuration ini files (they must be *.ini files). Normally, you should just override the specific option that you need to override. Note that any file that exists in `.ddev/php/` will be copied into `/etc/php/[version]/(cli|fpm)/conf.d`, so it's possible to replace files that already exist in the container. Common usage is to put custom overrides in a file called `my-php.ini`. Make sure you include the section header that goes with each item (like `[PHP]`)

One interesting implication of this behavior is that it's possible to disable extensions by replacing the configuration file that loads them. For instance, if you were to create an empty file at `.ddev/php/20-xdebug.ini`, it would replace the configuration that loads xdebug, which would cause xdebug to not be loaded!

To load the new configuration, just run a `ddev restart`.

An example file in .ddev/php/my-php.ini might look like this:
```
[PHP]
max_execution_time = 240;
```

## Providing custom mysql/MariaDB configuration (my.cnf)

You can provide additional PHP configuration for a project by creating a directory called `.ddev/mysql/` and adding any number of MySQL configuration files (these must have the suffix ".cnf"). These files will be automatically included when MySQL is started. Make sure that the section header is included in the file 

An example file in .ddev/mysql/no_utf8mb4.cnf might be:

```
[mysqld]
collation-server = utf8_general_ci
character-set-server = utf8
innodb_large_prefix=false
```

To load the new configuration, run `ddev restart`.

## Overriding default container images
The default container images provided by ddev are defined in the `config.yaml` file in the `.ddev` folder of your project. This means that _defining_ an alternative image for default services is as simple as changing the image definition in `config.yaml`. In practice, however, ddev currently has certain expectations and assumptions for what the web and database containers provide. At this time, it is recommended that the default container projects be referenced or used as a starting point for developing an alternative image. If you encounter difficulties integrating alternative images, please [file an issue and let us know](https://github.com/drud/ddev/issues/new).
