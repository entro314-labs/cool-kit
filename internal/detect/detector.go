package detect

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Detect attempts to detect the framework in the given directory
func Detect(dir string) (*FrameworkInfo, error) {
	// Check for Dockerfile first (highest priority)
	if fileExists(filepath.Join(dir, "Dockerfile")) {
		return detectDockerfile(dir)
	}

	// Check for Docker Compose
	if fileExists(filepath.Join(dir, "docker-compose.yml")) || fileExists(filepath.Join(dir, "docker-compose.yaml")) {
		return detectDockerCompose(dir)
	}

	// Check for package.json (Node.js projects)
	if fileExists(filepath.Join(dir, "package.json")) {
		return detectNodeProject(dir)
	}

	// Check for Laravel/PHP (composer.json)
	if fileExists(filepath.Join(dir, "composer.json")) {
		return detectPHPProject(dir)
	}

	// Check for Ruby/Rails (Gemfile)
	if fileExists(filepath.Join(dir, "Gemfile")) {
		return detectRubyProject(dir)
	}

	// Check for Elixir/Phoenix (mix.exs)
	if fileExists(filepath.Join(dir, "mix.exs")) {
		return detectElixirProject(dir)
	}

	// Check for Rust (Cargo.toml)
	if fileExists(filepath.Join(dir, "Cargo.toml")) {
		return detectRust(dir)
	}

	// Check for Hugo
	if fileExists(filepath.Join(dir, "hugo.toml")) || fileExists(filepath.Join(dir, "config.toml")) {
		if isHugoProject(dir) {
			return detectHugo(dir)
		}
	}

	// Check for Go
	if fileExists(filepath.Join(dir, "go.mod")) {
		return detectGo(dir)
	}

	// Check for Python
	if fileExists(filepath.Join(dir, "requirements.txt")) || fileExists(filepath.Join(dir, "pyproject.toml")) || fileExists(filepath.Join(dir, "Pipfile")) {
		return detectPythonProject(dir)
	}

	// Check for Deno (deno.json or deno.jsonc)
	if fileExists(filepath.Join(dir, "deno.json")) || fileExists(filepath.Join(dir, "deno.jsonc")) {
		return detectDeno(dir)
	}

	// Fallback to static site if index.html exists
	if fileExists(filepath.Join(dir, "index.html")) {
		return detectStatic(dir)
	}

	// Default to nixpacks with no specific framework
	return &FrameworkInfo{
		Name:      "Unknown",
		BuildPack: BuildPackNixpacks,
	}, nil
}

func detectNodeProject(dir string) (*FrameworkInfo, error) {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil, err
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		Scripts         map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	allDeps := make(map[string]string)
	for k, v := range pkg.Dependencies {
		allDeps[k] = v
	}
	for k, v := range pkg.DevDependencies {
		allDeps[k] = v
	}

	// Check for Bun lockfile first (changes install command)
	isBun := fileExists(filepath.Join(dir, "bun.lockb")) || fileExists(filepath.Join(dir, "bun.lock"))
	installCmd := "npm install"
	if isBun {
		installCmd = "bun install"
	} else if fileExists(filepath.Join(dir, "pnpm-lock.yaml")) {
		installCmd = "pnpm install"
	} else if fileExists(filepath.Join(dir, "yarn.lock")) {
		installCmd = "yarn install"
	}

	// Detect T3 Stack (before Next.js since it uses Next.js)
	if _, ok := allDeps["@trpc/server"]; ok {
		if _, ok := allDeps["next"]; ok {
			return &FrameworkInfo{
				Name:           "T3 Stack",
				BuildPack:      BuildPackNixpacks,
				InstallCommand: installCmd,
				BuildCommand:   "npm run build",
				StartCommand:   "npm start",
				Port:           "3000",
				IsStatic:       false,
			}, nil
		}
	}

	// Detect Next.js
	if _, ok := allDeps["next"]; ok {
		return &FrameworkInfo{
			Name:           "Next.js",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: installCmd,
			BuildCommand:   "npm run build",
			StartCommand:   "npm start",
			Port:           "3000",
			IsStatic:       false,
		}, nil
	}

	// Detect Remix
	if _, ok := allDeps["@remix-run/node"]; ok {
		return &FrameworkInfo{
			Name:           "Remix",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: installCmd,
			BuildCommand:   "npm run build",
			StartCommand:   "npm start",
			Port:           "3000",
			IsStatic:       false,
		}, nil
	}
	if _, ok := allDeps["@remix-run/react"]; ok {
		return &FrameworkInfo{
			Name:           "Remix",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: installCmd,
			BuildCommand:   "npm run build",
			StartCommand:   "npm start",
			Port:           "3000",
			IsStatic:       false,
		}, nil
	}

	// Detect NestJS
	if _, ok := allDeps["@nestjs/core"]; ok {
		return &FrameworkInfo{
			Name:           "NestJS",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: installCmd,
			BuildCommand:   "npm run build",
			StartCommand:   "node dist/main.js",
			Port:           "3000",
			IsStatic:       false,
		}, nil
	}

	// Detect AdonisJS
	if _, ok := allDeps["@adonisjs/core"]; ok {
		return &FrameworkInfo{
			Name:           "AdonisJS",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: installCmd,
			BuildCommand:   "node ace build --production",
			StartCommand:   "node build/server.js",
			Port:           "3333",
			IsStatic:       false,
		}, nil
	}

	// Detect Strapi
	if _, ok := allDeps["@strapi/strapi"]; ok {
		return &FrameworkInfo{
			Name:           "Strapi",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: installCmd,
			BuildCommand:   "npm run build",
			StartCommand:   "npm start",
			Port:           "1337",
			IsStatic:       false,
		}, nil
	}

	// Detect Astro
	if _, ok := allDeps["astro"]; ok {
		return &FrameworkInfo{
			Name:             "Astro",
			BuildPack:        BuildPackNixpacks,
			InstallCommand:   installCmd,
			BuildCommand:     "npm run build",
			PublishDirectory: "dist",
			Port:             "4321",
			IsStatic:         true,
		}, nil
	}

	// Detect Nuxt
	if _, ok := allDeps["nuxt"]; ok {
		return &FrameworkInfo{
			Name:           "Nuxt",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: installCmd,
			BuildCommand:   "npm run build",
			StartCommand:   "node .output/server/index.mjs",
			Port:           "3000",
			IsStatic:       false,
		}, nil
	}

	// Detect SvelteKit
	if _, ok := allDeps["@sveltejs/kit"]; ok {
		return &FrameworkInfo{
			Name:           "SvelteKit",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: installCmd,
			BuildCommand:   "npm run build",
			StartCommand:   "node build",
			Port:           "3000",
			IsStatic:       false,
		}, nil
	}

	// Detect SolidStart
	if _, ok := allDeps["@solidjs/start"]; ok {
		return &FrameworkInfo{
			Name:           "SolidStart",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: installCmd,
			BuildCommand:   "npm run build",
			StartCommand:   "npm start",
			Port:           "3000",
			IsStatic:       false,
		}, nil
	}

	// Detect Solid.js (static)
	if _, ok := allDeps["solid-js"]; ok {
		return &FrameworkInfo{
			Name:             "Solid.js",
			BuildPack:        BuildPackNixpacks,
			InstallCommand:   installCmd,
			BuildCommand:     "npm run build",
			PublishDirectory: "dist",
			Port:             "3000",
			IsStatic:         true,
		}, nil
	}

	// Detect Qwik
	if _, ok := allDeps["@builder.io/qwik"]; ok {
		return &FrameworkInfo{
			Name:           "Qwik",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: installCmd,
			BuildCommand:   "npm run build",
			StartCommand:   "npm run serve",
			Port:           "3000",
			IsStatic:       false,
		}, nil
	}

	// Detect Angular
	if _, ok := allDeps["@angular/core"]; ok {
		return &FrameworkInfo{
			Name:             "Angular",
			BuildPack:        BuildPackNixpacks,
			InstallCommand:   installCmd,
			BuildCommand:     "npm run build",
			PublishDirectory: "dist",
			Port:             "4200",
			IsStatic:         true,
		}, nil
	}

	// Detect Gatsby
	if _, ok := allDeps["gatsby"]; ok {
		return &FrameworkInfo{
			Name:             "Gatsby",
			BuildPack:        BuildPackNixpacks,
			InstallCommand:   installCmd,
			BuildCommand:     "gatsby build",
			PublishDirectory: "public",
			Port:             "9000",
			IsStatic:         true,
		}, nil
	}

	// Detect Vue.js
	if _, ok := allDeps["vue"]; ok {
		return &FrameworkInfo{
			Name:             "Vue.js",
			BuildPack:        BuildPackNixpacks,
			InstallCommand:   installCmd,
			BuildCommand:     "npm run build",
			PublishDirectory: "dist",
			Port:             "8080",
			IsStatic:         true,
		}, nil
	}

	// Detect Vite (generic)
	if _, ok := allDeps["vite"]; ok {
		return &FrameworkInfo{
			Name:             "Vite",
			BuildPack:        BuildPackNixpacks,
			InstallCommand:   installCmd,
			BuildCommand:     "npm run build",
			PublishDirectory: "dist",
			Port:             "5173",
			IsStatic:         true,
		}, nil
	}

	// Detect React (Create React App)
	if _, ok := allDeps["react-scripts"]; ok {
		return &FrameworkInfo{
			Name:             "Create React App",
			BuildPack:        BuildPackNixpacks,
			InstallCommand:   installCmd,
			BuildCommand:     "npm run build",
			PublishDirectory: "build",
			IsStatic:         true,
		}, nil
	}

	// Detect Express.js
	if _, ok := allDeps["express"]; ok {
		return &FrameworkInfo{
			Name:           "Express.js",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: installCmd,
			StartCommand:   "node index.js",
			Port:           "3000",
			IsStatic:       false,
		}, nil
	}

	// Detect Fastify
	if _, ok := allDeps["fastify"]; ok {
		return &FrameworkInfo{
			Name:           "Fastify",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: installCmd,
			StartCommand:   "node index.js",
			Port:           "3000",
			IsStatic:       false,
		}, nil
	}

	// Detect Hono
	if _, ok := allDeps["hono"]; ok {
		return &FrameworkInfo{
			Name:           "Hono",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: installCmd,
			StartCommand:   "node index.js",
			Port:           "3000",
			IsStatic:       false,
		}, nil
	}

	// Generic Node.js / Bun
	startCmd := ""
	if _, ok := pkg.Scripts["start"]; ok {
		startCmd = "npm start"
		if isBun {
			startCmd = "bun run start"
		}
	}
	buildCmd := ""
	if _, ok := pkg.Scripts["build"]; ok {
		buildCmd = "npm run build"
		if isBun {
			buildCmd = "bun run build"
		}
	}

	frameworkName := "Node.js"
	if isBun {
		frameworkName = "Bun"
	}

	return &FrameworkInfo{
		Name:           frameworkName,
		BuildPack:      BuildPackNixpacks,
		InstallCommand: installCmd,
		BuildCommand:   buildCmd,
		StartCommand:   startCmd,
		Port:           "3000",
		IsStatic:       false,
	}, nil
}

func detectHugo(dir string) (*FrameworkInfo, error) {
	return &FrameworkInfo{
		Name:             "Hugo",
		BuildPack:        BuildPackNixpacks,
		BuildCommand:     "hugo",
		PublishDirectory: "public",
		IsStatic:         true,
	}, nil
}

func isHugoProject(dir string) bool {
	// Check for typical Hugo directories
	return dirExists(filepath.Join(dir, "content")) ||
		dirExists(filepath.Join(dir, "themes")) ||
		dirExists(filepath.Join(dir, "layouts"))
}

func detectGo(dir string) (*FrameworkInfo, error) {
	return &FrameworkInfo{
		Name:         "Go",
		BuildPack:    BuildPackNixpacks,
		BuildCommand: "go build -o app",
		StartCommand: "./app",
		Port:         "8080",
		IsStatic:     false,
	}, nil
}

func detectPython(dir string) (*FrameworkInfo, error) {
	installCmd := "pip install -r requirements.txt"
	if fileExists(filepath.Join(dir, "pyproject.toml")) {
		installCmd = "pip install ."
	}

	return &FrameworkInfo{
		Name:           "Python",
		BuildPack:      BuildPackNixpacks,
		InstallCommand: installCmd,
		Port:           "8000",
		IsStatic:       false,
	}, nil
}

func detectStatic(dir string) (*FrameworkInfo, error) {
	return &FrameworkInfo{
		Name:             "Static Site",
		BuildPack:        BuildPackStatic,
		PublishDirectory: ".",
		Port:             "80",
		IsStatic:         true,
	}, nil
}

func detectDockerfile(dir string) (*FrameworkInfo, error) {
	return &FrameworkInfo{
		Name:      "Dockerfile",
		BuildPack: BuildPackDockerfile,
		Port:      "3000",
		IsStatic:  false,
	}, nil
}

func detectDockerCompose(dir string) (*FrameworkInfo, error) {
	return &FrameworkInfo{
		Name:      "Docker Compose",
		BuildPack: BuildPackDockerCompose,
		IsStatic:  false,
	}, nil
}

// detectPHPProject detects Laravel, Symfony, and other PHP frameworks
func detectPHPProject(dir string) (*FrameworkInfo, error) {
	data, err := os.ReadFile(filepath.Join(dir, "composer.json"))
	if err != nil {
		return &FrameworkInfo{
			Name:           "PHP",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: "composer install",
			Port:           "8000",
			IsStatic:       false,
		}, nil
	}

	var composer struct {
		Require map[string]string `json:"require"`
	}
	if err := json.Unmarshal(data, &composer); err != nil {
		return nil, err
	}

	// Detect Laravel
	if _, ok := composer.Require["laravel/framework"]; ok {
		return &FrameworkInfo{
			Name:           "Laravel",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: "composer install --no-dev --optimize-autoloader",
			BuildCommand:   "npm install && npm run build",
			StartCommand:   "php artisan serve --host=0.0.0.0 --port=8000",
			Port:           "8000",
			IsStatic:       false,
		}, nil
	}

	// Detect Symfony
	if _, ok := composer.Require["symfony/framework-bundle"]; ok {
		return &FrameworkInfo{
			Name:           "Symfony",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: "composer install --no-dev --optimize-autoloader",
			StartCommand:   "symfony server:start --port=8000",
			Port:           "8000",
			IsStatic:       false,
		}, nil
	}

	// Generic PHP
	return &FrameworkInfo{
		Name:           "PHP",
		BuildPack:      BuildPackNixpacks,
		InstallCommand: "composer install",
		Port:           "8000",
		IsStatic:       false,
	}, nil
}

// detectRubyProject detects Rails and other Ruby frameworks
func detectRubyProject(dir string) (*FrameworkInfo, error) {
	data, err := os.ReadFile(filepath.Join(dir, "Gemfile"))
	if err != nil {
		return &FrameworkInfo{
			Name:           "Ruby",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: "bundle install",
			Port:           "3000",
			IsStatic:       false,
		}, nil
	}

	content := string(data)

	// Detect Rails
	if strings.Contains(content, "rails") {
		return &FrameworkInfo{
			Name:           "Rails",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: "bundle install",
			BuildCommand:   "bundle exec rails assets:precompile",
			StartCommand:   "bundle exec rails server -b 0.0.0.0 -p 3000",
			Port:           "3000",
			IsStatic:       false,
		}, nil
	}

	// Detect Sinatra
	if strings.Contains(content, "sinatra") {
		return &FrameworkInfo{
			Name:           "Sinatra",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: "bundle install",
			StartCommand:   "ruby app.rb",
			Port:           "4567",
			IsStatic:       false,
		}, nil
	}

	// Generic Ruby
	return &FrameworkInfo{
		Name:           "Ruby",
		BuildPack:      BuildPackNixpacks,
		InstallCommand: "bundle install",
		Port:           "3000",
		IsStatic:       false,
	}, nil
}

// detectElixirProject detects Phoenix and other Elixir frameworks
func detectElixirProject(dir string) (*FrameworkInfo, error) {
	data, err := os.ReadFile(filepath.Join(dir, "mix.exs"))
	if err != nil {
		return &FrameworkInfo{
			Name:           "Elixir",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: "mix deps.get",
			StartCommand:   "mix run --no-halt",
			Port:           "4000",
			IsStatic:       false,
		}, nil
	}

	content := string(data)

	// Detect Phoenix
	if strings.Contains(content, ":phoenix") {
		return &FrameworkInfo{
			Name:           "Phoenix",
			BuildPack:      BuildPackNixpacks,
			InstallCommand: "mix deps.get && mix assets.deploy",
			BuildCommand:   "MIX_ENV=prod mix compile",
			StartCommand:   "MIX_ENV=prod mix phx.server",
			Port:           "4000",
			IsStatic:       false,
		}, nil
	}

	// Generic Elixir
	return &FrameworkInfo{
		Name:           "Elixir",
		BuildPack:      BuildPackNixpacks,
		InstallCommand: "mix deps.get",
		StartCommand:   "mix run --no-halt",
		Port:           "4000",
		IsStatic:       false,
	}, nil
}

// detectRust detects Rust projects
func detectRust(dir string) (*FrameworkInfo, error) {
	return &FrameworkInfo{
		Name:         "Rust",
		BuildPack:    BuildPackNixpacks,
		BuildCommand: "cargo build --release",
		StartCommand: "./target/release/$(basename $(pwd))",
		Port:         "8080",
		IsStatic:     false,
	}, nil
}

// detectPythonProject detects Django, Flask, FastAPI and other Python frameworks
func detectPythonProject(dir string) (*FrameworkInfo, error) {
	// Check requirements.txt for framework hints
	reqPath := filepath.Join(dir, "requirements.txt")
	if fileExists(reqPath) {
		data, err := os.ReadFile(reqPath)
		if err == nil {
			content := strings.ToLower(string(data))

			// Detect Django
			if strings.Contains(content, "django") {
				return &FrameworkInfo{
					Name:           "Django",
					BuildPack:      BuildPackNixpacks,
					InstallCommand: "pip install -r requirements.txt",
					BuildCommand:   "python manage.py collectstatic --noinput",
					StartCommand:   "gunicorn --bind 0.0.0.0:8000 $(basename $(pwd)).wsgi:application",
					Port:           "8000",
					IsStatic:       false,
				}, nil
			}

			// Detect FastAPI
			if strings.Contains(content, "fastapi") {
				return &FrameworkInfo{
					Name:           "FastAPI",
					BuildPack:      BuildPackNixpacks,
					InstallCommand: "pip install -r requirements.txt",
					StartCommand:   "uvicorn main:app --host 0.0.0.0 --port 8000",
					Port:           "8000",
					IsStatic:       false,
				}, nil
			}

			// Detect Flask
			if strings.Contains(content, "flask") {
				return &FrameworkInfo{
					Name:           "Flask",
					BuildPack:      BuildPackNixpacks,
					InstallCommand: "pip install -r requirements.txt",
					StartCommand:   "gunicorn --bind 0.0.0.0:5000 app:app",
					Port:           "5000",
					IsStatic:       false,
				}, nil
			}
		}
	}

	// Check for pyproject.toml
	installCmd := "pip install -r requirements.txt"
	if fileExists(filepath.Join(dir, "pyproject.toml")) {
		installCmd = "pip install ."
	} else if fileExists(filepath.Join(dir, "Pipfile")) {
		installCmd = "pipenv install"
	}

	return &FrameworkInfo{
		Name:           "Python",
		BuildPack:      BuildPackNixpacks,
		InstallCommand: installCmd,
		Port:           "8000",
		IsStatic:       false,
	}, nil
}

// detectDeno detects Deno projects
func detectDeno(dir string) (*FrameworkInfo, error) {
	// Check for Fresh framework (popular Deno framework)
	if fileExists(filepath.Join(dir, "fresh.gen.ts")) || dirExists(filepath.Join(dir, "islands")) {
		return &FrameworkInfo{
			Name:         "Deno Fresh",
			BuildPack:    BuildPackNixpacks,
			StartCommand: "deno task start",
			Port:         "8000",
			IsStatic:     false,
		}, nil
	}

	return &FrameworkInfo{
		Name:         "Deno",
		BuildPack:    BuildPackNixpacks,
		StartCommand: "deno run --allow-net --allow-read main.ts",
		Port:         "8000",
		IsStatic:     false,
	}, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
