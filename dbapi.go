package main

// проверяем vin номер
func (handler *IngosHandler) checkVin(vin string) (bool, error) {

	rows, err := handler.DB.Query("SELECT id FROM cars WHERE vin = ? LIMIT 1", vin)

	return rows.Next(), err

}

// получаем адрес кошелька по номеру автомобилю
func (handler *IngosHandler) getAddressByVin(vin string) (string, error) {

	row := handler.DB.QueryRow("SELECT id, address FROM cars WHERE vin = ?", vin)

	var id int
	var address string

	err := row.Scan(&id, &address)

	return address, err

}

// добавляем новый vin номер в базу данных
func (handler *IngosHandler) registrationVin(vin string, address string) error {

	result, err := handler.DB.Exec(
		"INSERT INTO cars (`vin`, `address`) VALUES (?, ?)",
		vin,
		address,
	)

	_, err = result.RowsAffected()

	return err
}
