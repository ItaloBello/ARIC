INSERT INTO aric.categorias (id, nome) VALUES
(1, 'Plástico'),
(2, 'Papelão'),
(3, 'Vidro'),
(4, 'Metal'),
(5, 'Orgânico'),
(6, 'Eletrônico'),
(7, 'Têxtil'),
(8, 'Madeira'),
(9, 'Borracha'),
(10, 'Outros'),
(11, 'Papel'),
(12, 'Energia Renovável'),
(13, 'Horta'),
(14, 'Alimentos'),
(15, 'Água')
ON CONFLICT (id) DO NOTHING;

-- Como os ids foram informados à mão, a sequence precisa ser ajustada,
-- senão o próximo INSERT sem id tenta usar o id 1 e dá conflito.
SELECT setval(pg_get_serial_sequence('aric.categorias', 'id'), (SELECT MAX(id) FROM aric.categorias));
