package modelo

type Rola struct {
     id              int
     idPerformer     int
     idAlbum         int
     artist          string
     album           string
     albumPath       string
     title           string
     path            string
     track           int
     totalTracks     int
     year            int
     genre           string
     comments        string
     composer        string
     originalArtist  string
     albumArtist     string
     disc            int
     totalDiscs      int
     images          [][]byte
}

// DatosRola representa los datos de una rola
// se intercambian mediante json
type DatosRola struct {
     ID      int    `json:"id"`
     AlbumID int    `json:"album_id"`
     Title   string `json:"title"`
     Artist  string `json:"artist"`
     Album   string `json:"album"`
     Year    int    `json:"year"`
     Genre   string `json:"genre"`
     Path    string `json:"path"`
}

// Datos regresa una representación de los datos de la rola
func (r Rola) Datos() DatosRola {
     return DatosRola{
     	    ID:      r.id,
	    AlbumID: r.idAlbum,
	    Title:   r.title,
	    Artist:  r.artist,
	    Album:   r.album,
	    Year:    r.year,
	    Genre:   r.genre,
	    Path:    r.path,
     }
}
