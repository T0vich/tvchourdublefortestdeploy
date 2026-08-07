-- Приводит схему к доменной модели, которая уехала вперёд.
--
-- Репозитории читают у products колонки title/image/price/location/status,
-- а у categories и wishlists — created_at. В 001 этих колонок нет, поэтому
-- без этой миграции падают и каталог, и справочник категорий.
--
-- Миграцию безопасно прогнать повторно: всё, что не имеет формы
-- IF NOT EXISTS, обёрнуто в проверку текущего состояния схемы.

-- ---------- products: переименование ----------

DO $$ BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'products' AND column_name = 'name'
    ) THEN
        ALTER TABLE products RENAME COLUMN name TO title;
    END IF;
END $$;

-- Функция триггера из 002 обращается к NEW.name, поэтому переопределяется
-- сразу после переименования — до любых INSERT и UPDATE по products,
-- иначе они падают с "record new has no field name".
-- Сам триггер products_search_trigger остаётся тот же.
CREATE OR REPLACE FUNCTION products_search_update()
RETURNS trigger
LANGUAGE plpgsql AS $$
DECLARE
    category_name TEXT;
BEGIN
    SELECT name INTO category_name
    FROM categories
    WHERE category_id = NEW.category_id;

    NEW.search_vector :=
        setweight(to_tsvector('simple',  coalesce(NEW.title, '')), 'A') ||
        setweight(to_tsvector('russian', coalesce(NEW.description, '')), 'B') ||
        setweight(to_tsvector('russian', coalesce(category_name, '')), 'C');

    RETURN NEW;
END $$;

-- ---------- products: новые колонки ----------

ALTER TABLE products
    ADD COLUMN IF NOT EXISTS image    TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS price    INTEGER      NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS location VARCHAR(255) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS status   VARCHAR(50)  NOT NULL DEFAULT 'active';

-- is_active был двоичным, статус — из четырёх значений;
-- снятые с публикации товары становятся archived.
DO $$ BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'products' AND column_name = 'is_active'
    ) THEN
        UPDATE products SET status = CASE WHEN is_active THEN 'active' ELSE 'archived' END;
        ALTER TABLE products DROP COLUMN is_active;
    END IF;
END $$;

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'products_status_check') THEN
        ALTER TABLE products
            ADD CONSTRAINT products_status_check
            CHECK (status IN ('active', 'reserved', 'exchanged', 'archived'));
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);

-- description участвует в поиске через similarity(), а он на NULL
-- возвращает NULL и роняет чтение оценки релевантности в float64.
UPDATE products SET description = '' WHERE description IS NULL;
ALTER TABLE products ALTER COLUMN description SET DEFAULT '';
ALTER TABLE products ALTER COLUMN description SET NOT NULL;

-- ---------- categories и wishlists ----------

ALTER TABLE categories
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE wishlists
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- ---------- перестройка поискового вектора ----------

-- Холостой UPDATE поднимает триггер и пересобирает search_vector
-- под новым именем колонки для всех уже лежащих товаров.
UPDATE products SET title = title;
