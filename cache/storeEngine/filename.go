package storeEngine

import (
	"fmt"
	"os"
)

const (
	LogFile = iota
	DBLockFile
	TableFile
	DescriptorFile
	CurrentFile
	TempFile
	InfoLogFile
)

func makeFileName(dbname string, number uint64, suffix string) string {
	return fmt.Sprintf(
		"%s/%06llu.%s",
		dbname,
		number,
		suffix,
	)
}

func LogFileName(dbname string, number uint64) string {
	return makeFileName(dbname, number, "log")
}

func TableFileName(dbname string, number uint64) string {
	return makeFileName(dbname, number, "ldb")
}

func SstTableFileName(dbname string, number uint64) string {
	return makeFileName(dbname, number, "sst")
}

func CurrentFileName(dbname string) string {
	return dbname + "/CURRENT"
}

func DescriptorFileName(dbname string, number uint64) string {
	return fmt.Sprintf("%s/MANIFEST-%06llu",
		dbname,
		number)
}

func LockFileName(dbname string) string {
	return dbname + "/LOCK"
}

func TempFileName(dbname string, number uint64) string {
	return makeFileName(dbname, number, "dbtmp")
}

func InfoLogFileName(dbname string) string {
	return dbname + "/LOG"
}

func OldInfoLogFileName(dbname string) string {
	return dbname + "/LOG.old"
}

// Owned filenames have the form:
//    dbname/CURRENT
//    dbname/LOCK
//    dbname/LOG
//    dbname/LOG.old
//    dbname/MANIFEST-[0-9]+
//    dbname/[0-9]+.(log|sst|ldb)
// func parseFileName(const std::string& fname,
//                    uint64_t* number,
//                    FileType* type) bool {
//   Slice rest(fname);
//   if (rest == "CURRENT") {
//     *number = 0;
//     *type = kCurrentFile;
//   } else if (rest == "LOCK") {
//     *number = 0;
//     *type = kDBLockFile;
//   } else if (rest == "LOG" || rest == "LOG.old") {
//     *number = 0;
//     *type = kInfoLogFile;
//   } else if (rest.starts_with("MANIFEST-")) {
//     rest.remove_prefix(strlen("MANIFEST-"));
//     uint64_t num;
//     if (!ConsumeDecimalNumber(&rest, &num)) {
//       return false;
//     }
//     if (!rest.empty()) {
//       return false;
//     }
//     *type = kDescriptorFile;
//     *number = num;
//   } else {
//     // Avoid strtoull() to keep filename format independent of the
//     // current locale
//     uint64_t num;
//     if (!ConsumeDecimalNumber(&rest, &num)) {
//       return false;
//     }
//     Slice suffix = rest;
//     if (suffix == Slice(".log")) {
//       *type = kLogFile;
//     } else if (suffix == Slice(".sst") || suffix == Slice(".ldb")) {
//       *type = kTableFile;
//     } else if (suffix == Slice(".dbtmp")) {
//       *type = kTempFile;
//     } else {
//       return false;
//     }
//     *number = num;
//   }
//   return true;
// }

func setCurrentFile(dbname string, number uint64) error {
	manifest := descriptorFileName(dbname, number)
	manifest = manifest[len(dbname)+1:]
	tmp := tempFileName(dbname, number)
	// write to temp file
	// writetoFile(manifest+"\n", tmp)
	err := os.Rename(tmp, currentFileName(dbname))
	if err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}
