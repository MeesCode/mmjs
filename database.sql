-- running this will empty your old database and create a new one

DROP SCHEMA IF EXISTS mmjs;
CREATE SCHEMA IF NOT EXISTS mmjs;
USE mmjs;

DROP TABLE IF EXISTS mmjs.Folders;
CREATE TABLE IF NOT EXISTS mmjs.Folders (
  FolderID int NOT NULL AUTO_INCREMENT,
  Path varchar(191) NOT NULL UNIQUE,
  ParentID int DEFAULT NULL REFERENCES Folders(FolderID),
  PRIMARY KEY (FolderID)
) ENGINE=InnoDB;

DROP TABLE IF EXISTS mmjs.Tracks;
CREATE TABLE IF NOT EXISTS mmjs.Tracks (
  TrackID int NOT NULL AUTO_INCREMENT,
  Path varchar(191) NOT NULL UNIQUE,
  FolderID int NOT NULL,
  Title varchar(191) DEFAULT NULL,
	Album varchar(191) DEFAULT NULL,
	Artist varchar(191) DEFAULT NULL,
	Genre varchar(191) DEFAULT NULL,
	Year int DEFAULT NULL,
  Plays int DEFAULT 0,
  PRIMARY KEY (TrackID),
  FOREIGN KEY (FolderID) REFERENCES Folders(FolderID)
) ENGINE=InnoDB;

DROP TABLE IF EXISTS mmjs.Playlists;
CREATE TABLE IF NOT EXISTS mmjs.Playlists (
  PlaylistID int NOT NULL AUTO_INCREMENT,
  Name varchar(191) NOT NULL UNIQUE,
  PRIMARY KEY (PlaylistID)
) ENGINE=InnoDB;

DROP TABLE IF EXISTS mmjs.PlaylistEntries;
CREATE TABLE IF NOT EXISTS mmjs.PlaylistEntries (
  PlaylistEntryID int NOT NULL AUTO_INCREMENT,
  TrackID int NOT NULL,
  PlaylistID int NOT NULL,
  FOREIGN KEY (TrackID) REFERENCES Tracks(TrackID),
  FOREIGN KEY (PlaylistID) REFERENCES Playlists(PlaylistID),
  PRIMARY KEY (PlaylistEntryID)
) ENGINE=InnoDB;