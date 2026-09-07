-- name: ListDeals :many
SELECT
    best_deals.item_id,
    items.name AS item_name,
    best_deals.buy_platform_id,
    buy_platform.name AS buy_platform_name,
    best_deals.sell_platform_id,
    sell_platform.name AS sell_platform_name,
    best_deals.buy_price,
    best_deals.sell_price
FROM best_deals
JOIN items ON items.id = best_deals.item_id
JOIN platforms AS buy_platform ON buy_platform.id = best_deals.buy_platform_id
JOIN platforms AS sell_platform ON sell_platform.id = best_deals.sell_platform_id
ORDER BY best_deals.profit DESC, best_deals.item_id;
