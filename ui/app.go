package ui

func (c *Console) mainMenu() {
	for {
		c.println("")
		c.println("=== CLOUDFLARED SETUP CLI ===")
		c.println("[1] Autentikasi & Setup")
		c.println("[2] Manajemen Kredensial")
		c.println("[3] Observability & Monitoring")
		c.println("[4] Orkestrasi")
		c.println("[5] Pemeliharaan")
		c.println("[0] Keluar")
		choice := c.readChoice()
		if c.eof {
			return
		}
		switch choice {
		case "":
			continue
		case "1":
			c.authMenu()
		case "2":
			c.credentialsMenu()
		case "3":
			c.monitoringMenu()
		case "4":
			c.orchestrationMenu()
		case "5":
			c.maintenanceMenu()
		case "0", "q":
			c.info("Sampai jumpa.")
			return
		default:
			c.fail("Pilihan tidak valid")
		}
	}
}
