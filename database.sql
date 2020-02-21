-- running this will empty your old database and create a new one

DROP SCHEMA IF EXISTS mmjs;
CREATE SCHEMA IF NOT EXISTS mmjs;
USE mmjs;

DROP TABLE IF EXISTS mmjs.Folders;
CREATE TABLE IF NOT EXISTS mmjs.Folders (
  FolderID int NOT NULL AUTO_INCREMENT,
  Path VARCHAR(255) NOT NULL UNIQUE,
  ParentID int DEFAULT NULL REFERENCES Folders(FolderID),
  PRIMARY KEY (FolderID)
) ENGINE=InnoDB;

DROP TABLE IF EXISTS mmjs.Tracks;
CREATE TABLE IF NOT EXISTS mmjs.Tracks (
  TrackID int NOT NULL AUTO_INCREMENT,
  Path VARCHAR(255) NOT NULL UNIQUE,
  FolderID int NOT NULL,
  Title varchar(255) DEFAULT NULL,
	Album varchar(255) DEFAULT NULL,
	Artist varchar(255) DEFAULT NULL,
	Genre varchar(255) DEFAULT NULL,
	Year int DEFAULT NULL,
  PRIMARY KEY (TrackID),
  FOREIGN KEY (FolderID) REFERENCES Folders(FolderID)
) ENGINE=InnoDB;

DROP TABLE IF EXISTS mmjs.Playlist;
CREATE TABLE IF NOT EXISTS mmjs.Playlist (
  PlaylistID int NOT NULL AUTO_INCREMENT,
  Name VARCHAR(255) NOT NULL UNIQUE,
  PRIMARY KEY (PlaylistID)
) ENGINE=InnoDB;

DROP TABLE IF EXISTS mmjs.PlaylistEntry;
CREATE TABLE IF NOT EXISTS mmjs.PlaylistEntry (
  PlaylistEntryID int NOT NULL AUTO_INCREMENT,
  TrackID int NOT NULL,
  PlaylistID int NOT NULL,
  FOREIGN KEY (TrackID) REFERENCES Tracks(TrackID),
  FOREIGN KEY (PlaylistID) REFERENCES Playlist(PlaylistID),
  PRIMARY KEY (PlaylistEntryID)
) ENGINE=InnoDB;