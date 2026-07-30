This Git repository should contain a Go application that compiles into a single (per-platform), self-contained executable containing application logic, web-based user interface and API documentation.

Do a "git pull" at the start of operations to make sure the local repository is up-to-date.

The application should be self-hostable by a home user. If run on a home desktop machine (Windows, MacOS, Linux, Raspberry Pi) the executable should start up the back-end server to listen on an available non-privileged port, and should then start up an available web browser to point at the user interface served by that back-end. In this operation mode, the application should only be available to the local web browser, it shouldn't accept traffic from the wider network, and user authentication should be by simply reading the username from the local environment.

The application should use Go's "embed" package to host all needed file resources.

The application should be able to operate on a machine without internet access. The user interface should be web-based, constructed using React, with the React library files served by the application itself.

The application should also work in a multi-user environment when hosted on a server (probably inside a minimal Docker container) behind a reverse proxy server (Pangolin / Traefik, Cloudflare Tunnel). In this case, the proxy server will handle user authentication and pass in a header value to give the current authenticated username.

The application should serve its user interface from the root "/" endpoint.

The application should serve its application logic from the "/api" endpoint.

The application should serve auto-generated OpenAPI documentation from the "docs/api" endpoint.

~/www is a live web server folder. Once files have been compiled, please copy them to the ~/www/go-app/ folder for distribution. Please generate an "index.html" file that includes the contents of the README.md file and contains links to download the executables for each platform.

Add any new or modified files to the Git repository and do a "git push".
