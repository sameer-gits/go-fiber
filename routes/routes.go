package routes

import (
    "fmt"
    "log"
    "os"

    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/session"
    "github.com/gofiber/fiber/v2/utils"
    "github.com/gofiber/storage/memory/v2"
    "github.com/joho/godotenv"
    "github.com/markbates/goth"
    "github.com/markbates/goth/providers/github"
    "github.com/shareed2k/goth_fiber"
)

func Routes() {

    app := fiber.New()

    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    githubKey := os.Getenv("GITHUB_KEY")
    githubSecret := os.Getenv("GITHUB_SECRET")

    // optional config
    config := session.Config{
        Expiration:     time.Hour * 24 * 30,
        Storage:        memory.New(), 
        KeyLookup:      "header:session_name",
        CookieDomain:   "",
        CookiePath:     "/",
        CookieSecure:   false,
        CookieHTTPOnly: true, // Should always be enabled
        CookieSameSite: "Lax",
        KeyGenerator:   utils.UUIDv4,
    }

    // create session handler
    sessions := session.New(config)

    goth_fiber.SessionStore = sessions

    goth.UseProviders(
        github.New(githubKey, githubSecret, "http://localhost:8080/auth/github/callback"),
    )

    app.Get("/auth/:provider", goth_fiber.BeginAuthHandler)

    app.Get("/auth/:provider/callback", func(ctx *fiber.Ctx) error {
        user, err := goth_fiber.CompleteUserAuth(ctx)
        if err != nil {
            log.Fatal(err)
        }

        sess, err := sessions.Get(ctx)
        if err != nil {
            return err
        }

        sess.Set("email", user.Email)


        return ctx.SendString(user.Email)
    })
    app.Get("/logout/:provider", func(ctx *fiber.Ctx) error {
        if err := goth_fiber.Logout(ctx); err != nil {
            log.Fatal(err)
        }

        return ctx.SendString("logout")
    })

    // Home route
    app.Get("/", func(ctx *fiber.Ctx) error {
        sess, err := sessions.Get(ctx)
        if err != nil {
            return err
        }
        email := sess.Get("email")
        return ctx.SendString(fmt.Sprintf("User Email: %v", email))
    })

    log.Fatal(app.Listen(":8080"))

}
