# One-Time Secret Share

One-Time Secret Share is a secure application for sharing one-time viewable messages using unique URLs. Each message can only be accessed once, ensuring privacy and security.

## Project Structure

The project directory is organized as follows:


```
one-time-secret-share/
├── html
    ├── header.html # Common header template
    ├── index.html # Home page template
    ├── layout.html # Base layout template
    ├── message.html # Template for displaying messages
    ├── note.html # Template for displaying a single note
    ├── notfound.html # Template for 404 error page
    └── success.html # Template for success messages
├── handler.go # HTTP request handlers
├── helpers.go # Helper functions for error handling and template rendering
├── models.go # Data models for the application
├── main.go # Entry point for the application
├── Dockerfile # Docker configuration for building and running the application
├── docker-compose.yml # Docker Compose configuration for services
└── README.md # This documentation file
```




## Features

- **Single-Use Links**: Each message is accessible only once via a unique URL.
- **Enhanced Security**: Messages are securely stored and destructible.
- **User-Friendly Interface**: Simple API for message generation and access.

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) (version 1.22.5 or later)
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

### Installation

1. **Clone the repository**:
    ```bash
    git clone https://github.com/KhanbalaRashidov/one-time-secret-share.git
    ```

2. **Navigate into the project directory**:
    ```bash
    cd one-time-secret-share
    ```

3. **Build and run the application using Docker Compose**:
    ```bash
    docker-compose up -d
    ```

### Main Components

1. **`main.go`**: Sets up the web server and Redis cache.
   - **Setup**: Initializes Redis cache and starts the HTTP server.
   - **Example**:
     ```go
     func main() {
         // Load settings
         port := os.Getenv("PORT")
         if len(port) == 0 {
             port = "3000"
         }
         addr := ":" + port

         // Initialize Redis cache
         redisOptions, err := redis.ParseURL(os.Getenv("REDIS_URL"))
         if err != nil {
             panic(err)
         }
         redisClient := redis.NewClient(redisOptions)
         defer redisClient.Close()
         redisCache := cache.New(&cache.Options{
             Redis: redisClient,
         })
         server := &Server{
             BaseURL:    fmt.Sprintf("http://localhost:%s", port),
             RedisCache: redisCache,
         }

         // Start web server
         fmt.Printf("Starting web server on %s\n", addr)
         err = http.ListenAndServe(addr, server)
         if err != nil {
             panic(err)
         }
     }
     ```

2. **`handler.go`**: Manages HTTP requests.
   - **POST `/`**: Generates a one-time link for the message.
     - **Example**:
       ```go
       func (s *Server) handlePOST(w http.ResponseWriter, r *http.Request) {
           err := r.ParseForm()
           if err != nil {
               s.badRequest(w, r, http.StatusBadRequest, "Invalid form data posted.")
               return
           }
           message := r.PostForm.Get("message")
           key := uuid.NewString()
           note := &Note{
               Data:     []byte(message),
               Destruct: true,
           }
           err = s.RedisCache.Set(&cache.Item{
               Ctx:     r.Context(),
               Key:     key,
               Value:   note,
               TTL:     time.Hour * 24 * 365,
           })
           if err != nil {
               s.serverError(w, r)
               return
           }
           noteURL := fmt.Sprintf("%s/%s", s.BaseURL, key)
           s.renderTemplate(w, r, struct{ NoteURL string }{NoteURL: noteURL}, "layout", "html/layout.html", "html/success.html")
       }
       ```

   - **GET `/{id}`**: Retrieves a one-time message.
     - **Example**:
       ```go
       func (s *Server) handleGET(w http.ResponseWriter, r *http.Request) {
           path := r.URL.Path
           if path == "/" {
               s.renderTemplate(w, r, nil, "layout", "html/layout.html", "html/index.html")
               return
           }
           noteID := strings.TrimPrefix(path, "/")
           ctx := r.Context()
           note := &Note{}
           err := s.RedisCache.GetSkippingLocalCache(ctx, noteID, note)
           if err != nil {
               s.notFound(w, r, "Note Not Found", fmt.Sprintf("Note with ID %s does not exist.", noteID))
               return
           }
           if note.Destruct {
               s.RedisCache.Delete(ctx, noteID)
           }
           s.renderTemplate(w, r, struct{ Title string; NoteContent template.HTML }{Title: "Note", NoteContent: template.HTML(string(note.Data))}, "layout", "html/layout.html", "html/note.html")
       }
       ```

3. **`helpers.go`**: Contains helper functions for rendering templates and error handling.
   - **Example**:
     ```go
     func (s *Server) renderTemplate(w http.ResponseWriter, r *http.Request, data interface{}, name string, files ...string) {
         t, err := template.ParseFiles(files...)
         if err != nil {
             http.Error(w, "Template parsing error: "+err.Error(), http.StatusInternalServerError)
             return
         }
         err = t.ExecuteTemplate(w, name, data)
         if err != nil {
             http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
         }
     }
     ```

4. **`models.go`**: Defines data models used by the application.
   - **Example**:
     ```go
     type Note struct {
         Data     []byte
         Destruct bool
     }

     type Server struct {
         BaseURL    string
         RedisCache *cache.Cache
     }
     ```

### Configuration

- **PORT**: Port on which the application listens. Default is `3000`.
- **REDIS_URL**: URL for the Redis database. Default is `redis://:@localhost:6379/1`.
- **BASE_URL**: Base URL for generating one-time links. Default is `http://localhost:PORT`.

### Docker Integration

- **Dockerfile**: Defines the steps to build a Docker image for the application.
  - **Example**:
    ```dockerfile
    FROM golang:alpine AS build-env
    RUN apk add git
    ARG version=0.0.0
    WORKDIR /app
    COPY . .
    RUN go build -o /go/bin ./...

    FROM alpine:latest
    ARG version
    ENV APP_VERSION=$version
    WORKDIR /app
    COPY --from=build-env /go/bin/one-time-secret-share .
    COPY --from=build-env /app/html ./html
    ENTRYPOINT ["./one-time-secret-share"]
    ```

- **docker-compose.yml**: Defines services for the application and Redis database.
  - **Example**:
    ```yaml
    version: "3.9"

    services:
      db:
        image: redis:latest
        ports:
          - "6379:6379"

      web:
        build: ./
        environment:
          - PORT=3000
          - REDIS_URL=redis://:@db:6379/1
        ports:
          - "3000:3000"
        depends_on:
          - db
    ```

## Usage

1. **Generate a One-Time Link**:
   - Make a POST request to `/` with the `message` field in the form data.

2. **Access the Message**:
   - Use the provided link to view the message. The link will be valid for a single view only.

## Contributing

Contributions are welcome! If you have suggestions, improvements, or bug fixes, please open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).

## Contact

For any inquiries or feedback, please contact [Khanbala Rashidov](https://github.com/KhanbalaRashidov).

---

Happy coding! 🚀
