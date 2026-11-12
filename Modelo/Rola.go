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