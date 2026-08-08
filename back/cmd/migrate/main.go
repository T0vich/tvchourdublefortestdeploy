// Команда migrate применяет SQL-файлы к базе, указанной в DATABASE_URL.
//
// В docker-compose схема накатывается автоматически через
// docker-entrypoint-initdb.d, но во внешних окружениях (Neon и любой другой
// managed Postgres) такого механизма нет — там миграции нужно запускать явно.
//
//	go run ./cmd/migrate                       # все файлы из infrastructure/migrations
//	go run ./cmd/migrate path/to/one.sql ...   # только указанные файлы, в заданном порядке
//
// Файл разбивается на отдельные операторы и выполняется по одному: так ошибка
// указывает на конкретный оператор, а не на файл целиком.
package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

const (
	defaultMigrationsDir = "infrastructure/migrations"

	// statementTimeout ограничивает ожидание ответа по одному оператору.
	statementTimeout = 30 * time.Second
)

// verbose включается через MIGRATE_VERBOSE=1 и печатает каждый оператор перед
// выполнением — нужно, когда миграция где-то встаёт и надо понять, где именно.
var verbose = os.Getenv("MIGRATE_VERBOSE") != ""

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL не задан")
	}

	files, err := resolveFiles(os.Args[1:])
	if err != nil {
		log.Fatalf("не удалось собрать список файлов: %s", err)
	}
	if len(files) == 0 {
		log.Fatal("не найдено ни одного .sql файла")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("не удалось подключиться к базе: %s", err)
	}
	defer func() { _ = conn.Close(context.Background()) }()

	for _, file := range files {
		body, readErr := os.ReadFile(file) //nolint:gosec // путь задаёт оператор, это CLI для миграций
		if readErr != nil {
			log.Fatalf("не удалось прочитать %s: %s", file, readErr)
		}

		statements := splitStatements(string(body))
		for i, statement := range statements {
			if verbose {
				log.Printf("  [%d/%d] %s", i+1, len(statements), preview(statement))
			}

			// Отдельный дедлайн на каждый оператор: если ответ сервера не
			// доходит из-за сетевой заминки, лучше упасть с указанием
			// конкретного оператора, чем висеть до общего таймаута.
			stmtCtx, stmtCancel := context.WithTimeout(ctx, statementTimeout)
			_, execErr := conn.Exec(stmtCtx, statement)
			stmtCancel()

			if execErr != nil {
				log.Fatalf("%s, оператор %d из %d (%s): %s",
					filepath.Base(file), i+1, len(statements), preview(statement), execErr)
			}
		}

		log.Printf("применён %s (%d операторов)", filepath.Base(file), len(statements))
	}

	log.Printf("готово, файлов применено: %d", len(files))
}

// resolveFiles возвращает либо явно переданные пути, либо всё содержимое
// каталога миграций по умолчанию, отсортированное по имени.
func resolveFiles(args []string) ([]string, error) {
	if len(args) > 0 {
		return args, nil
	}

	entries, err := os.ReadDir(defaultMigrationsDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		files = append(files, filepath.Join(defaultMigrationsDir, entry.Name()))
	}
	sort.Strings(files)

	return files, nil
}

// preview сжимает оператор в одну короткую строку для лога.
func preview(statement string) string {
	flat := strings.Join(strings.Fields(statement), " ")
	if len(flat) > 70 {
		return flat[:70] + "..."
	}

	return flat
}

// splitStatements режет SQL на операторы по точке с запятой, не трогая те,
// что находятся внутри строк, комментариев или долларовых кавычек. Последнее
// обязательно: тестовые данные обёрнуты в блок DO $$ ... $$, внутри которого
// точек с запятой больше, чем снаружи.
//
// Сканирование идёт по байтам, а не по рунам: все значимые разделители — ASCII,
// а многобайтовые последовательности UTF-8 никогда не содержат ASCII-байтов,
// так что кириллица в комментариях и строках проходит насквозь без искажений.
func splitStatements(sql string) []string {
	var (
		statements []string
		current    strings.Builder
	)

	flush := func() {
		if trimmed := strings.TrimSpace(current.String()); trimmed != "" {
			statements = append(statements, trimmed)
		}
		current.Reset()
	}

	for i := 0; i < len(sql); {
		rest := sql[i:]

		switch {
		case strings.HasPrefix(rest, "--"):
			if end := strings.IndexByte(rest, '\n'); end != -1 {
				i += end + 1
			} else {
				i = len(sql)
			}
			current.WriteByte('\n')

		case strings.HasPrefix(rest, "/*"):
			if end := strings.Index(rest[2:], "*/"); end != -1 {
				i += 2 + end + 2
			} else {
				i = len(sql)
			}
			current.WriteByte(' ')

		case sql[i] == '\'':
			end := strings.IndexByte(rest[1:], '\'')
			if end == -1 {
				current.WriteString(rest)
				i = len(sql)
				break
			}
			current.WriteString(rest[:end+2])
			i += end + 2

		case sql[i] == '$':
			tag := dollarTag(rest)
			if tag == "" {
				current.WriteByte('$')
				i++
				break
			}
			closing := strings.Index(rest[len(tag):], tag)
			if closing == -1 {
				current.WriteString(rest)
				i = len(sql)
				break
			}
			end := len(tag) + closing + len(tag)
			current.WriteString(rest[:end])
			i += end

		case sql[i] == ';':
			flush()
			i++

		default:
			current.WriteByte(sql[i])
			i++
		}
	}
	flush()

	return statements
}

// dollarTag распознаёт открывающую долларовую кавычку ($$ или $tag$) в начале
// строки и возвращает её целиком; если это просто символ $, возвращает "".
func dollarTag(s string) string {
	for i := 1; i < len(s); i++ {
		if s[i] == '$' {
			return s[:i+1]
		}
		isWordByte := s[i] == '_' ||
			(s[i] >= 'a' && s[i] <= 'z') ||
			(s[i] >= 'A' && s[i] <= 'Z') ||
			(s[i] >= '0' && s[i] <= '9')
		if !isWordByte {
			return ""
		}
	}

	return ""
}
