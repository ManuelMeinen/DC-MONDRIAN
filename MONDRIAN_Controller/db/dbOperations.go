package db

import (
	"context"
	"controller/types"
	"database/sql"
	"fmt"
	"net"
	_ "github.com/mattn/go-sqlite3"
)

var maxZoneID = types.ZoneID(1<<24 - 1)

/* Getters */

// GetAllSites returns all sites stored in the backend
func (b *Backend) GetAllSites() (types.Sites, error) {

	stmt := `SELECT tp_address, name FROM sites`

	rows, err := b.db.Query(stmt)
	if err != nil {
		return nil, err
	}

	var sites types.Sites
	var tp string
	var name string
	for rows.Next() {
		err = rows.Scan(&tp, &name)
		if err != nil {
			return nil, err
		}
		sites = append(sites, &types.Site{TPAddr: tp, Name: name})
	}
	return sites, nil
}

// GetAllZones returns all zones stored in the backend
func (b *Backend) GetAllZones() (types.Zones, error) {

	stmt := `SELECT id, name FROM zones`

	rows, err := b.db.Query(stmt)
	if err != nil {
		return nil, err
	}

	var zones types.Zones
	var id int
	var name string
	for rows.Next() {
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, err
		}
		zones = append(zones, &types.Zone{ID: types.ZoneID(id), Name: name})
	}
	return zones, nil
}

// GetAllSubnets returns all subnets stored in the backend
func (b *Backend) GetAllSubnets() (types.Subnets, error) {

	stmt := `SELECT net_ip, net_mask, zone, tp_address 
	FROM subnets`

	rows, err := b.db.Query(stmt)
	if err != nil {
		return nil, err
	}

	var nets types.Subnets
	var ip []byte
	var mask []byte
	var zone types.ZoneID
	var tp string
	for rows.Next() {
		err = rows.Scan(&ip, &mask, &zone, &tp) 
		if err != nil {
			return nil, err
		}
		nets = append(nets, &types.Subnet{IPNet: net.IPNet{IP: ip, Mask: mask}, ZoneID: zone, TPAddr: tp})
	}
	return nets, nil
}

// GetSubnets returns all subnets stored in the backend relevant for ZTP with address tpAddr
func (b *Backend) GetSubnets(tpAddr string) (types.Subnets, error) { //TODO fix this such that wildcards are accepted
	stmt := `WITH tp_zones
	AS (SELECT DISTINCT zone 
		FROM   subnets 
		WHERE  tp_address = ?), 
	possible_dests 
	AS (SELECT DISTINCT dest
		FROM   transitions 
		WHERE  (src IN tp_zones OR src = 0 OR src IS NULL) AND dest NOT IN tp_zones), 
	possible_src 
	AS (SELECT DISTINCT src
		FROM  transitions 
		WHERE  (dest IN tp_zones OR dest = 0 OR dest IS NULL) AND src NOT IN tp_zones),
	dest_wildcard
	AS (SELECT DISTINCT src
		FROM transitions 
		WHERE dest IS NULL OR dest = 0),
	src_wildcard
	AS (SELECT DISTINCT dest
		FROM transitions
		WHERE src IS NULL OR src = 0), 
	wildcard_zone_count_src
	AS(	SELECT DISTINCT count(dest) 
		FROM src_wildcard
		WHERE dest IN tp_zones),
	wildcard_zone_count_dest
	AS(	SELECT DISTINCT count(src) 
		FROM dest_wildcard
		WHERE src IN tp_zones)
	SELECT net_ip, net_mask, zone, tp_address 
	FROM   subnets 
	WHERE  zone IN tp_zones OR zone IN possible_dests
	   OR zone IN possible_src
	   OR ((SELECT count(*) FROM tp_zones JOIN dest_wildcard WHERE zone = src) > 0)
	   OR ((SELECT count(*) FROM tp_zones JOIN src_wildcard WHERE zone = dest) > 0)`

	rows, err := b.db.Query(stmt, tpAddr)
	if err != nil {
		return nil, err
	}

	var nets types.Subnets
	var ip []byte
	var mask []byte
	var zone types.ZoneID
	var tp string
	for rows.Next() {
		err = rows.Scan(&ip, &mask, &zone, &tp)
		if err != nil {
			return nil, err
		}
		nets = append(nets, &types.Subnet{IPNet: net.IPNet{IP: ip, Mask: mask}, ZoneID: zone, TPAddr: tp})
	}
	return nets, nil
}

// GetAllTransitions returns all transitions stored in the backend
func (b *Backend) GetAllTransitions() ([]types.Transition, error) {
	
	stmt := `SELECT policyID, src, dest, srcPort, destPort, proto, action 
	FROM   transitions`

	rows, err := b.db.Query(stmt)
	if err != nil {
		return nil, err
	}
	
	var t types.Transition
	var transitions []types.Transition
	var policyID int
	var src sql.NullInt32
	var dest sql.NullInt32
	var srcPort sql.NullInt32
	var destPort sql.NullInt32
	var proto sql.NullString
	var action string
	for rows.Next() {
		err = rows.Scan(&policyID, &src, &dest, &srcPort, &destPort, &proto, &action)
		if err != nil {
			return nil, err
		}
		t.PolicyID = uint(policyID)
		t.Src = types.ZoneID(src.Int32)
		t.Dest = types.ZoneID(types.GetInt(dest))
		t.SrcPort = uint(types.GetInt(srcPort))
		t.DestPort = uint(types.GetInt(destPort))
		t.Proto = types.GetString(proto)
		t.Action = action
		
		transitions = append(transitions, t)
	}
	return transitions, nil
}

// GetTransitions returns all transitions of a given TP stored in the backend
func (b *Backend) GetTransitions(tpAddr string) ([]types.Transition, error) {
	
	stmt := `WITH relevant_zones 
	AS (SELECT DISTINCT zone 
		FROM   subnets 
		WHERE  tp_address = ?) 
	SELECT policyID, src, dest, srcPort, destPort, proto, action
	FROM   transitions 
	WHERE  src IN relevant_zones 
	   OR dest IN relevant_zones
	   OR src IS NULL
	   OR dest IS NULL`

	rows, err := b.db.Query(stmt, tpAddr)
	if err != nil {
		return nil, err
	}

	var transitions []types.Transition
	var t types.Transition
	var policyID int
	var src sql.NullInt32
	var dest sql.NullInt32
	var srcPort sql.NullInt32
	var destPort sql.NullInt32
	var proto sql.NullString
	var action string
	for rows.Next() {
		err = rows.Scan(&policyID, &src, &dest, &srcPort, &destPort, &proto, &action)
		if err != nil {
			return nil, err
		}
		t.PolicyID = uint(policyID)
		t.Src = types.ZoneID(src.Int32)
		t.Dest = types.ZoneID(types.GetInt(dest))
		t.SrcPort = uint(types.GetInt(srcPort))
		t.DestPort = uint(types.GetInt(destPort))
		t.Proto = types.GetString(proto)
		t.Action = action
		
		transitions = append(transitions, t)
	}
	return transitions, nil
}

/* Insertions */

// InsertSites inserts sites into the Backend
func (b *Backend) InsertSites(sites types.Sites) error {
	stmt := `INSERT INTO sites (tp_address, name) VALUES (?, ?)`

	// do insertion in a transaction to ensure atomicity
	tx, err := b.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	for _, site := range sites {
		//TODO(mmeinen): Check if TPAddr is in a valid format
		_, err = tx.Exec(stmt, site.TPAddr, site.Name)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}

// InsertZones inserts zones into the Backend
func (b *Backend) InsertZones(zones types.Zones) error {
	stmt := `INSERT INTO zones (id, name) VALUES (?, ?)`

	// do insertion in a transaction to ensure atomicity
	tx, err := b.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	for _, zone := range zones {
		// check zoneID is not too big
		if zone.ID > maxZoneID {
			tx.Rollback()
			return fmt.Errorf("zone ID must be at most %d", maxZoneID)
		}
		_, err := tx.Exec(stmt, zone.ID, zone.Name)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

// InsertSubnets inserts subnets into the Backend
func (b *Backend) InsertSubnets(subnets types.Subnets) error {
	stmt := `INSERT INTO Subnets (zone, net_ip, net_mask, tp_address) VALUES (?, ?, ?, ?)`

	// do insertion in a transaction to ensure atomicity
	tx, err := b.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	for _, subnet := range subnets {
		if subnet.ZoneID > maxZoneID {
			tx.Rollback()
			return fmt.Errorf("zone ID must be at most %d", maxZoneID)
		}
		// TODO(mmeinen): check if TPAddr is a valid address
		_, err = tx.Exec(stmt, subnet.ZoneID, subnet.IPNet.IP, subnet.IPNet.Mask, subnet.TPAddr)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}

// InsertTransitions inserts zone transitions into the Backend
func (b *Backend) InsertTransitions(transitions types.Transitions) error {
	stmt := `INSERT INTO transitions VALUES (?, ?, ?, ?, ?, ?, ?)`

	// do insertion in a transaction to ensure atomicity
	tx, err := b.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	for _, transition := range transitions {
		var src sql.NullInt32
		if transition.Src == 0 {
			src = sql.NullInt32{Int32: 0, Valid: false}
		}else{
			src = sql.NullInt32{Int32: int32(transition.Src), Valid:true}
		}
		var dest sql.NullInt32
		if transition.Dest == 0 {
			dest = sql.NullInt32{Int32: 0, Valid: false}
		}else{
			dest = sql.NullInt32{Int32: int32(transition.Dest), Valid:true}
		}
		_, err = tx.Exec(stmt, transition.PolicyID, src, dest, transition.SrcPort, transition.DestPort, transition.Proto, transition.Action)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

/* Deletions */

// DeleteSites deletes branch sites from the Backend
func (b *Backend) DeleteSites(sites types.Sites) error {
	stmt := `DELETE FROM sites WHERE tp_address = ?`

	// do deletion in a transaction to ensure atomicity
	tx, err := b.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	for _, site := range sites {
		_, err = tx.Exec(stmt, site.TPAddr)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}

// DeleteZones deletes zones from the Backend
func (b *Backend) DeleteZones(zones types.Zones) error {
	stmt := `DELETE FROM zones WHERE id = ?`

	// do deletions in a transaction to ensure atomicity
	tx, err := b.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	for _, zone := range zones {
		// check zoneID is not too big
		if zone.ID > maxZoneID {
			tx.Rollback()
			return fmt.Errorf("zone ID must be at most %d", maxZoneID)
		}
		_, err := tx.Exec(stmt, zone.ID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}

// DeleteSubnets delete subnets from the Backend
func (b *Backend) DeleteSubnets(subnets types.Subnets) error {
	stmt := `DELETE FROM subnets WHERE net_ip = ? AND net_mask = ?`

	// do deletions in a transaction to ensure atomicity
	tx, err := b.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	for _, subnet := range subnets {
		_, err = tx.Exec(stmt, subnet.IPNet.IP, subnet.IPNet.Mask)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}

// DeleteTransitions delete zone transitions into the Backend
func (b *Backend) DeleteTransitions(transitions types.Transitions) error {
	stmt := `DELETE FROM Transitions WHERE policyID = ?`

	// do insertion in a transaction to ensure atomicity
	tx, err := b.db.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}

	for _, transition := range transitions {
		_, err = tx.Exec(stmt, transition.PolicyID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}