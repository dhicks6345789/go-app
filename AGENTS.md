This Git repository should contain a Go application that compiles into a single self-contained executable containing application logic, web-based user interface and API documentation.

The application should be self-hostable by a home user. If run on a home desktop machine (Windows, MacOS, Linux, Raspberry Pi) the executable should start up the back-end server to listen on an available non-privileged port, and should then start up an available web browser to point at the user interface served by that back-end. In this operation mode, the application should only be available to the local web browser, it shouldn't accept traffic from the wider network, and user authentication should be by simply reading the username from the local environment.

The application should use Go's "embed" package to host all needed file resources.

The application should be able to operate on a machine without internet access. The user interface should be web-based, constructed using React, with the React library files served by the application itself.

Parts include:
- A self-contained HTTP-only server. HTTPS will be handled by an external reverse proxy server (Pangolin / Traefik, Cloudflare Tunnel). That server will also handle user authentication and pass in a header value to give the current authenticated username.
- 
- At the root "/" endpoint, the HTTP server should serve files embedded inside the Go executable itself using Go's embeded filesystem functionality.
- At the "/api" endpoint, the HTTP server should serve auto-generated OpenAPI documentation.
