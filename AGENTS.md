This Git repository should contain a Go application that compiles into a single (per-platform), self-contained executable containing application logic, web-based user interface and API documentation.

This Git repository is hosted on Github at: https://github.com/dhicks6345789/go-app

The project homepage is at: https://users.sansay.co.uk/d.b.hicks/go-app

Do a "git pull" at the start of operations to make sure the local repository is up-to-date.

The application should be self-hostable by a home user. If run on a home desktop machine (Windows, MacOS, Linux, Raspberry Pi) the executable should start up the back-end server to listen on an available non-privileged port, and should then start up an available web browser to point at the user interface served by that back-end. In this operation mode, the application should only be available to the local web browser, it shouldn't accept traffic from the wider network, and user authentication should be by simply reading the username from the local environment.

The application should use Go's "embed" package to host all needed file resources.

The application should be able to operate on a machine without internet access. The user interface should be web-based, constructed using React, with the React library files served by the application itself.

The application should also work in a multi-user environment when hosted on a server (probably inside a minimal Docker container) behind a reverse proxy server (Pangolin / Traefik, Cloudflare Tunnel). In this case, the proxy server will handle user authentication and pass in a header value to give the current authenticated username.

The application should serve its user interface from the root "/" endpoint.

The application should serve its application logic from the "/api" endpoint.

The application should serve auto-generated OpenAPI documentation from the "docs/api" endpoint.

Create an "index.html" file that includes the same information as README.md to explain and document the project to the public. Include links to live, downloadable executables for each platform. Include a link in index.html to the Github project. Include index.html in the Git repository, but exclude the compiled executable files.

Make sure the "docs/api.html" file is generated.

The index.html file, compiled executables and the contents of the "docs" folder will be copied to the live website by an external bash script.

Add any new or modified files to the Git repository and do a "git push".
