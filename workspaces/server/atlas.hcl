data "external_schema" "gorm" {
  program = [
    "go",
    "run",
    "-mod=mod",
    "./cmd/atlas/loader/loader.go"
  ]
}

env "gorm" {
  src = data.external_schema.gorm.url
  migration {
    dir = "file://migrations"
  }
}