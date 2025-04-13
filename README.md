# Origin

**Origin** is a progressive Go web framework designed to simplify building robust applications with modern idioms. It uses Go generics, reflection, GORM, and Iris to build scalable RESTful APIs with minimal boilerplate.

## Features

- **Generic ORM & Repository Layer:**  
  Reuse the same CRUD logic for any model with a strongly typed generic repository.
  
- **Dynamic Parameter Generation:**  
  Create, Update, and Filter parameter structures are generated dynamically via reflection. These parameter structs exclude base fields and support partial updates.

- **Multilingual Content Support:**  
  Built-in helper functions (like `ExtractContent`, `GetAllContentsWithUpdated`, and `IsContentModel`) automatically merge content models by language identifier.

- **Structured Logging:**  
  Integrated with [Logrus](https://github.com/sirupsen/logrus) for detailed, structured logs during CRUD operations and transactions.

- **Association Preloading & Full-Save:**  
  Easily preload associations (such as "Contents") on Get operations and use GORM’s FullSaveAssociations mode during Create/Update.

- **Iris Integration:**  
  Leverage the [Iris](https://github.com/kataras/iris) web framework to register routes and build RESTful APIs quickly.

## Installation

Origin requires Go 1.18+ (for generics) and uses Go modules. To install:

1. Clone the repository:

   ```bash
   git clone https://github.com/MuhmdHsn313/origin.git
   cd origin
   ```

2. Tidy your modules:

   ```bash
   go mod tidy
   ```

## Usage Example

Below is a complete example showing how to define your model, set up the repository, service, and API handlers using Iris:

```go
package main

import (
    "github.com/MuhmdHsn313/origin/orm"
    "github.com/MuhmdHsn313/origin/repository"
    "github.com/MuhmdHsn313/origin/service"
    "github.com/kataras/iris/v12"
    "github.com/sirupsen/logrus"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

// Blog defines the main blog model.
type Blog struct {
    orm.Model

    // Contents holds multilingual content for the blog.
    Contents []BlogContent `json:"contents" gorm:"foreignKey:BlogID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    
    // Owner of the blog.
    Owner string `json:"owner"`
    
    // IsPublished indicates if the blog is published.
    IsPublished bool `json:"is_published" gorm:"default:false;index"`
}

// BlogContent defines the content model for a blog (must implement IContentModel).
type BlogContent struct {
    orm.ContentModel

    // Content holds the text of the blog content.
    Content string `json:"content"`
    
    // BlogID is the foreign key linking to the Blog.
    BlogID  uint   `json:"blog_id"`
}

func main() {
    // Open a GORM SQLite database connection.
    db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    if err != nil {
        panic("failed to connect database")
    }

    // Migrate the schema for Blog and BlogContent models.
    db.AutoMigrate(&Blog{}, &BlogContent{})

    // Initialize a logger.
    logger := logrus.New()

    // Create an Iris server.
    irisServer := iris.Default()

    // Create the service engine and repository for Blog.
    eng := service.CreateEngine[Blog]()
    repo := repository.NewGenericRepository[Blog](db, logger)

    // Register API routes under the /api path.
    api := irisServer.Party("/api")
    blogService := service.NewModelService[Blog](eng, repo)
    service.RegisterHandler[Blog](api, blogService)

    // Start the Iris server.
    irisServer.Listen(":8080")
}
```

### Code Highlights

- **Model Definition:**  
  The `Blog` model embeds `orm.Model` and includes a `Contents` slice of type `BlogContent`. The `BlogContent` type embeds `orm.ContentModel` and implements the `IContentModel` interface (required for multilingual content operations).

- **Repository & Service Setup:**  
  The repository (`GenericRepository[Blog]`) and service (`NewModelService[Blog]`) use generics so that CRUD operations, preloading, and associations are handled consistently for any model type.

- **API Handler Registration:**  
  The service registers its HTTP handlers onto the Iris router under the `/api` path.

## Contributing

Contributions are welcome! Please open issues, submit pull requests, or discuss enhancements on the [GitHub repository](https://github.com/MuhmdHsn313/origin).

## License

Origin is open-source software licensed under the GNU GENERAL PUBLIC License.
