package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Post represents a blog post
type Post struct {
	ID        string    `json:"id,omitempty" bson:"_id,omitempty"`
	Title     string    `json:"title" bson:"title"`
	Content   string    `json:"content" bson:"content"`
	AuthorID  string    `json:"author_id,omitempty" bson:"author_id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty" bson:"created_at,omitempty"`
}
type User struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

var (
	client     *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
)

func main() {
	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	var err error
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	db = client.Database("UserBlog") // Use the "blog" database
	collection = db.Collection("posts")

	// Set up Gin router
	router := gin.Default()

	router.POST("/auth/signup", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Insert user data into MongoDB
		_, err := db.Collection("users").InsertOne(context.Background(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user": user})
	})

	// Routes
	router.POST("/posts", createPost)
	router.GET("/posts", getAllPosts)
	router.GET("/posts/:id", getPostByID)
	router.PUT("/posts/:id", updatePost)
	router.DELETE("/posts/:id", deletePost)

	// Start server
	if err := router.Run(":8000"); err != nil {
		log.Fatal(err)
	}

}

// createPost creates a new blog post
func createPost(c *gin.Context) {
	var post Post
	if err := c.BindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	post.ID = uuid.New().String()
	post.CreatedAt = time.Now()

	// Insert the post into the database
	if _, err := collection.InsertOne(context.Background(), post); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	c.JSON(http.StatusCreated, post)
}

// getAllPosts retrieves all blog posts
func getAllPosts(c *gin.Context) {
	var posts []Post
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve posts"})
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var post Post
		if err := cursor.Decode(&post); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve posts"})
			return
		}
		posts = append(posts, post)
	}

	c.JSON(http.StatusOK, posts)
}

// getPostByID retrieves a single blog post by ID
func getPostByID(c *gin.Context) {
	id := c.Param("id")
	var post Post
	if err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&post); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}
	c.JSON(http.StatusOK, post)
}

// updatePost updates a blog post by ID
func updatePost(c *gin.Context) {
	id := c.Param("id")
	var post Post
	if err := c.BindJSON(&post); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	post.ID = id
	_, err := collection.ReplaceOne(context.Background(), bson.M{"_id": id}, post)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post"})
		return
	}
	c.Status(http.StatusOK)
}

// deletePost deletes a blog post by ID
func deletePost(c *gin.Context) {
	id := c.Param("id")
	result, err := collection.DeleteOne(context.Background(), bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		return
	}
	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}
	c.Status(http.StatusOK)

}
