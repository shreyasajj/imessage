// mautrix-imessage - A Matrix-iMessage puppeting bridge.
// Copyright (C) 2021 Tulir Asokan
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package database

import (
	"database/sql"

	log "maunium.net/go/maulogger/v2"

	"maunium.net/go/mautrix/id"
)

type PortalQuery struct {
	db  *Database
	log log.Logger
}

func (pq *PortalQuery) New() *Portal {
	return &Portal{
		db:  pq.db,
		log: pq.log,
	}
}

func (pq *PortalQuery) GetAll() []*Portal {
	return pq.getAll("SELECT * FROM portal")
}

func (pq *PortalQuery) GetByGUID(guid string) *Portal {
	return pq.get("SELECT * FROM portal WHERE guid=$1", guid)
}

func (pq *PortalQuery) GetByMXID(mxid id.RoomID) *Portal {
	return pq.get("SELECT * FROM portal WHERE mxid=$1", mxid)
}

func (pq *PortalQuery) FindPrivateChats() []*Portal {
	// TODO make sure this is right
	return pq.getAll("SELECT * FROM portal WHERE guid LIKE '%;-;%'")
}

func (pq *PortalQuery) getAll(query string, args ...interface{}) (portals []*Portal) {
	rows, err := pq.db.Query(query, args...)
	if err != nil || rows == nil {
		return nil
	}
	defer rows.Close()
	for rows.Next() {
		portals = append(portals, pq.New().Scan(rows))
	}
	return
}

func (pq *PortalQuery) get(query string, args ...interface{}) *Portal {
	row := pq.db.QueryRow(query, args...)
	if row == nil {
		return nil
	}
	return pq.New().Scan(row)
}

type Portal struct {
	db  *Database
	log log.Logger

	GUID string
	MXID id.RoomID

	Name      string
	Avatar    string
	AvatarURL id.ContentURI
	Encrypted bool
}

func (portal *Portal) Scan(row Scannable) *Portal {
	var mxid, avatarURL sql.NullString
	err := row.Scan(&portal.GUID, &mxid, &portal.Name, &portal.Avatar, &avatarURL, &portal.Encrypted)
	if err != nil {
		if err != sql.ErrNoRows {
			portal.log.Errorln("Database scan failed:", err)
		}
		return nil
	}
	portal.MXID = id.RoomID(mxid.String)
	portal.AvatarURL, _ = id.ParseContentURI(avatarURL.String)
	return portal
}

func (portal *Portal) mxidPtr() *id.RoomID {
	if len(portal.MXID) > 0 {
		return &portal.MXID
	}
	return nil
}

func (portal *Portal) Insert() {
	_, err := portal.db.Exec("INSERT INTO portal (guid, mxid, name, avatar, avatar_url, encrypted) VALUES ($1, $2, $3, $4, $5, $6)",
		portal.GUID, portal.mxidPtr(), portal.Name, portal.Avatar, portal.AvatarURL.String(), portal.Encrypted)
	if err != nil {
		portal.log.Warnfln("Failed to insert %s: %v", portal.GUID, err)
	}
}

func (portal *Portal) Update() {
	var mxid *id.RoomID
	if len(portal.MXID) > 0 {
		mxid = &portal.MXID
	}
	_, err := portal.db.Exec("UPDATE portal SET mxid=$1, name=$2, avatar=$3, avatar_url=$4, encrypted=$5 WHERE guid=$6",
		mxid, portal.Name, portal.Avatar, portal.AvatarURL.String(), portal.Encrypted, portal.GUID)
	if err != nil {
		portal.log.Warnfln("Failed to update %s: %v", portal.GUID, err)
	}
}

func (portal *Portal) Delete() {
	_, err := portal.db.Exec("DELETE FROM portal WHERE guid=$1", portal.GUID)
	if err != nil {
		portal.log.Warnfln("Failed to delete %s: %v", portal.GUID, err)
	}
}
