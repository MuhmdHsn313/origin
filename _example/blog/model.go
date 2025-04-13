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
	BlogID uint `json:"blog_id"`
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
