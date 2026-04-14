-- name: InsertCompanyProduct :one
INSERT INTO company_products(company_id, product_id, kind, has_stock, for_sale, cost_per_unit, profit_margin, sale_price)
    VALUES (sqlc.arg('CompanyID'), sqlc.arg('ProductID'), sqlc.arg('Kind'), sqlc.narg('HasStock'), sqlc.narg('ForSale'), sqlc.arg('CostPerUnit'), sqlc.narg('ProfitMargin'), sqlc.arg('SalePrice'))
RETURNING
    *;

-- name: GetCompanyProductByID :one
SELECT
    cp.id,
    cp.company_id,
    cp.product_id,
    cp.kind,
    cp.has_stock,
    cp.for_sale,
    cp.cost_per_unit,
    cp.profit_margin,
    cp.sale_price,
    cp.created_at,
    cp.updated_at,
    cp.deleted_at
FROM
    company_products cp
WHERE
    cp.id = sqlc.arg('ID')
    AND cp.company_id = sqlc.arg('CompanyID')
    AND cp.deleted_at IS NULL
LIMIT 1;

-- name: UpdateCompanyProduct :one
UPDATE
    company_products
SET
    kind = coalesce(sqlc.narg('Kind'), kind),
    has_stock = coalesce(sqlc.narg('HasStock'), has_stock),
    for_sale = coalesce(sqlc.narg('ForSale'), for_sale),
    cost_per_unit = coalesce(sqlc.narg('CostPerUnit'), cost_per_unit),
    profit_margin = coalesce(sqlc.narg('ProfitMargin'), profit_margin),
    sale_price = coalesce(sqlc.narg('SalePrice'), sale_price),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND company_id = sqlc.arg('CompanyID')
    AND deleted_at IS NULL
RETURNING
    *;

-- name: DeleteCompanyProduct :one
UPDATE
    company_products
SET
    deleted_at = now(),
    updated_at = now()
WHERE
    id = sqlc.arg('ID')
    AND company_id = sqlc.arg('CompanyID')
    AND deleted_at IS NULL
RETURNING
    *;

-- name: ListCompanyProducts :many
SELECT
    cp.id,
    cp.company_id,
    cp.product_id,
    cp.kind,
    cp.has_stock,
    cp.for_sale,
    cp.cost_per_unit,
    cp.profit_margin,
    cp.sale_price,
    cp.created_at,
    cp.updated_at,
    cp.deleted_at,
    p.name AS product_name,
    p.description AS product_description,
    p.batch_number AS product_batch_number,
    p.expiration_date AS product_expiration_date,
    p.quantity AS product_quantity,
    p.image_url AS product_image_url
FROM
    company_products cp
    JOIN products p ON cp.product_id = p.id
WHERE
    cp.company_id = sqlc.arg('CompanyID')
    AND cp.deleted_at IS NULL
    AND p.deleted_at IS NULL
ORDER BY
    cp.created_at DESC
LIMIT sqlc.arg('Limit')
OFFSET sqlc.arg('Offset');

