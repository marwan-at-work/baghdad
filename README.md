# Baghdad

Scalable CI for micro-services.
--

### Intro

What if you could build 10, 50, 100, Docker images simultaneously at the same speed of building one image?

What if all of those images are in one repo, because your project is micro-services oriented?

What if you have more than one repo, with many Dockerfiles inside each one?

Baghdad runs within your swarm cluster, builds, versions, and deploys your micro-services, along with itself.

Baghdad also leverages every node within your cluster to parallelize Docker image builds.

### Highlights

- Build and deploy multiple repos simultaneously.
- Build and deploy multiple services within the same repository, also simultaneously.
- Specify multiple Dockerfiles within the same repository, multi-stage builds supported.
- Automatically tag your repo and create releases on github.
- Automatically extract files from built images and push them as artifacts to github releases.
- Monitor your builds from the CLI.
- Monitor your app logs from the CLI.
- Check Pull Requests' build status with a simple UI.

### Status

- Baghdad is not yet battle tested, and APIs might change until 1.0

### Usage (simple, no config.)

1. Make sure you have a working Baghdad ecosystem, and that your Github Webhook is properly setup (see Deploy To Production).

2. Create a Dockerfile in your root folder.

3. Enjoy : )

4. All your commits to master will trigger an image build, and push to your image repository (Dockerhub or others), and will also create a pre-release on github. Pull Requests will trigger a build but without any tags/pushes to repos.

5. To deploy any of these releases run: `bag deploy --env <env> --branch master --tag <tag>`

### Usage (detailed, with config)

1. For builds, you will need a `baghdad.toml` in your root directory. Refer to the [full example](https://github.com/marwan-at-work/baghdad/blob/master/example-baghdad.toml) for documentation.

2. For deploys, you will need a `stack-compose.yml` in your root directory. This is the docker-compose file that will be included in the `docker stack deploy --compose-file...` command. Baghdad runs on Baghdad, so you can see its own [stack-compose.yml](https://github.com/marwan-at-work/baghdad/blob/master/stack-compose.yml) for reference. There are two things to note about this file:

    - You do not need to include a tag in the image property, as Baghdad will do that for you based on the deploy tag. For example say you have a stack-compose file with the following config:
        ```
        version: "3.2"
        services:
          web:
            image: my-org/my-web-image
        ```

        if you do `bag deploy --tag master-1.0.0-Build.33`, Baghdad will deploy your projec with the following stack file:

        ```
        version: "3.2"
        services:
          web:
            image: my-org/my-web-image:master-1.0.0-Build.33
        ```

    - Baghdad runs your stack within Traefik, so you do not need to specify those properties in your stack-compose.yml, Baghdad will do it for you.

### Deploy to Production

1. Make sure you have at least one EC2 instance, with docker installed (v17.06.0-ce-rc2) and Swarm Mode enabled: `docker swarm init`

2. Create a `traefik.toml` file with the following config:

    ```
    [entryPoints]
    [entryPoints.http]
    address = ":80"
    [entryPoints.http.redirect]
    entryPoint = "https"
    [entryPoints.https]
    address = ":443"
    [entryPoints.https.tls]
    [[entryPoints.https.tls.certificates]]
    CertFile = "/ssl/cert.pem"
    KeyFile = "/ssl/key.pem”
    ````

    this assumes you have key/cert for https. If you don't, feel free to tweak the config above to only use https.

3. Deploy Traefik to the swarm

    ```
    docker service create \
        --name traefik \
        --detach=false \
        --constraint=node.role==manager \
        --publish 80:80 --publish 8080:8080 --publish 443:443 \
        --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
        --mount type=bind,source=$PWD/traefik.toml,target=/etc/traefik/traefik.toml \
        --mount type=bind,source=$PWD/cert.pem,target=/ssl/cert.pem \
        --mount type=bind,source=$PWD/key.pem,target=/ssl/key.pem \
        --network traefik-net \
        traefik:v1.3.0-rc3 \
        --docker \
        --docker.swarmmode \
        --docker.domain=traefik \
        --docker.watch \
        --logLevel=DEBUG \
        --web
    ```

4. Navigate to your Route53 or DNS provider and route `*.YOUR_DOMAIN.COM` to the IP address where Traefik is deployed.

5. From the Baghdad repo, run `bag generate stack --env prod --host YOUR_DOMAIN.COM --version <pick-ur-release-tag>`

6. Copy the stdout into the EC2 working direcory under `stack.yml`

7. Make sure to add the following secret file to the swarm under the secret name `baghdad-vars`:

```
ADMIN_TOKEN=<your-github-token> # will give Baghdad access to git clone your repo in order to build it.
DOCKER_ORG=<DOCKER_ORG_NAME> # where built docker images get puhsed.
DOCKER_AUTH_USER=<DOCKER_AUTH_USER>
DOCKER_AUTH_PASS=<DOCKER_AUTH_PASS>
BAGHDAD_DOMAIN_NAME=<DOMAIN_NAME> # example.com NOT www.example.com.
```

8. run `docker stack deploy --compose-file stack.yml baghdad_prod`

9. Create a github hook that points to `https://master-prod-baghdad-api.YOUR_DOMAIN.COM/hooks/github`. The passcode is `baghdad` (required). Make sure to disable SSL, if it's a self signed certificate.

10. Say a prayer ☪☮ℰ✡☥☯✝

### Development

Baghdad consists of many services. You can instantiate all of them or any of them. The easiest way to do that,
is through the `docker-compose.yml` file (not to be confused with `stack-compose.yml`). Each service defined in that file is geared for development mode. You only need to make sure you have a `.env` file in the root direcory with the same settings as `baghdad-vars` in the Deploy to Production section.
Make sure to have `BAGHDAD_DOMAIN_NAME=localhost`

Make sure your your local docker daemon is in swarm mode.

Also, make sure you deploy traeffik as a swarm service with a different port than 80, and that's where you can test deployed projects:

```
docker service create \
    --name traefik \
    --detach=false \
    --constraint=node.role==manager \
    --publish 3456:80 --publish 9090:8080 \
    --mount type=bind,source=/var/run/docker.sock,target=/var/run/docker.sock \
    --network traefik-net \
    traefik:v1.3.0-rc3 \
    --docker \
    --docker.swarmmode \
    --docker.domain=traefik \
    --docker.watch \
    --logLevel=DEBUG \
    --web
```

Note that that I'm binding `3456` to the main router, and `9090` for the traefik web UI, because the Baghdad rabbitmq binds `8080`.

### Roadmap

- [ ] E2E tests.
- [ ] Automated memory monitoring & recovery.
- [ ] Document architecture.
