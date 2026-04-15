env "dev" {
  // Define the URL of the database which is managed
  url = "postgresql://genpos:genpos@localhost:2028/genpos_dev?sslmode=disable"

  // Define the URL of the Dev Database for this environment
  dev = "docker://postgres/15/dev?search_path=public"

  // Define migration directory configuration
  migration {
    dir = "file://migrations"
  }
}

env "test" {
  // Define the URL of the database which is managed
  url = "postgresql://genpos:genpos@localhost:2028/genpos_test?sslmode=disable"

  // Define the URL of the Dev Database for this environment
  dev = "docker://postgres/15/dev?search_path=public"

  // Define migration directory configuration
  migration {
    dir = "file://migrations"
  }

  diff {
    skip {
      drop_schema = true
      drop_table = true
    }
  }
}
