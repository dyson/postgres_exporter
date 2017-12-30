package collector

import (
	"context"
	"database/sql"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Subsystem
	statDatabaseSubsystem = "stat_database"
	// Scrape query
	statDatabaseQuery = `SELECT datname, numbackends, tup_returned, tup_fetched, tup_inserted, tup_updated, tup_deleted,
							   xact_commit, xact_rollback, blks_read, blks_hit, conflicts, deadlocks,
							   temp_files, temp_bytes
						FROM pg_stat_database`
)

type statDatabaseCollector struct {
	numbackends  *prometheus.Desc
	tupReturned  *prometheus.Desc
	tupFetched   *prometheus.Desc
	tupInserted  *prometheus.Desc
	tupUpdated   *prometheus.Desc
	tupDeleted   *prometheus.Desc
	xactCommit   *prometheus.Desc
	xactRollback *prometheus.Desc
	blksRead     *prometheus.Desc
	blksHit      *prometheus.Desc
	conflicts    *prometheus.Desc
	deadlocks    *prometheus.Desc
	tempFiles    *prometheus.Desc
	tempBytes    *prometheus.Desc
}

func init() {
	registerCollector("stat_database", defaultEnabled, NewStatDatabaseCollector)
}

// NewStatDatabaseCollector returns a new Collector exposing postgres pg_stat_database view
// The Statistics Collector
// PostgreSQL's statistics collector is a subsystem that supports collection and reporting of information about
// server activity. Presently, the collector can count accesses to tables and indexes in both disk-block and
// individual-row terms. It also tracks the total number of rows in each table, and information about vacuum
// and analyze actions for each table. It can also count calls to user-defined functions and the total time
// spent in each one.
// https://www.postgresql.org/docs/9.4/static/monitoring-stats.html#PG-STAT-DATABASE-VIEW
func NewStatDatabaseCollector() (Collector, error) {
	return &statDatabaseCollector{
		numbackends: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, statDatabaseSubsystem, "numbackends"),
			"Number of backends currently connected to this database. This is the only column in this"+
				" view that returns a value reflecting current state; all other columns return the accumulated"+
				" values since the last reset.",
			[]string{"datname"},
			nil,
		),
		tupReturned: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, statDatabaseSubsystem, "tup_returned_total"),
			"Number of rows returned by queries in this database",
			[]string{"datname"},
			nil,
		),
		tupFetched: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, statDatabaseSubsystem, "tup_fetched_total"),
			"Number of rows fetched by queries in this database",
			[]string{"datname"},
			nil,
		),
		tupInserted: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, statDatabaseSubsystem, "tup_inserted_total"),
			"Number of rows inserted by queries in this database",
			[]string{"datname"},
			nil,
		),
		tupUpdated: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, statDatabaseSubsystem, "tup_updated_total"),
			"Number of rows updated by queries in this database",
			[]string{"datname"},
			nil,
		),
		tupDeleted: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, statDatabaseSubsystem, "tup_deleted_total"),
			"Number of rows deleted by queries in this database",
			[]string{"datname"},
			nil,
		),
		xactCommit: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, statDatabaseSubsystem, "xact_commit_total"),
			"Number of transactions in this database that have been committed",
			[]string{"datname"},
			nil,
		),
		xactRollback: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, statDatabaseSubsystem, "xact_rollback_total"),
			"Number of transactions in this database that have been rolled back",
			[]string{"datname"},
			nil,
		),
		blksRead: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, statDatabaseSubsystem, "blks_read_total"),
			"Number of disk blocks read in this database",
			[]string{"datname"},
			nil,
		),
		blksHit: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, statDatabaseSubsystem, "blks_hit_total"),
			"Number of times disk blocks were found already in the buffer cache, so that a read was not necessary"+
				" (this only includes hits in the PostgreSQL buffer cache, not the operating system's file system cache)",
			[]string{"datname"},
			nil,
		),
		conflicts: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, statDatabaseSubsystem, "conflicts_total"),
			"Number of queries canceled due to conflicts with recovery in this database."+
				" (Conflicts occur only on standby servers; see pg_stat_database_conflicts for details.)",
			[]string{"datname"},
			nil,
		),
		deadlocks: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, statDatabaseSubsystem, "deadlocks_total"),
			"Number of deadlocks detected in this database",
			[]string{"datname"},
			nil,
		),
		tempFiles: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, statDatabaseSubsystem, "temp_files_total"),
			"Number of temporary files created by queries in this database. All temporary files are counted,"+
				" regardless of why the temporary file was created (e.g., sorting or hashing), and regardless of "+
				" the log_temp_files setting.",
			[]string{"datname"},
			nil,
		),
		tempBytes: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, statDatabaseSubsystem, "temp_bytes_total"),
			"Total amount of data written to temporary files by queries in this database. All temporary files"+
				" are counted, regardless of why the temporary file was created, and regardless of the log_temp_files setting.",
			[]string{"datname"},
			nil,
		),
	}, nil
}

func (c *statDatabaseCollector) Update(ctx context.Context, db *sql.DB, ch chan<- prometheus.Metric) error {
	rows, err := db.QueryContext(ctx, statDatabaseQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	var datname string
	var numbackends, tupReturned, tupFetched, tupInserted, tupUpdated, tupDeleted, xactCommit, xactRollback,
		blksRead, blksHit, conflicts, deadlocks, tempFiles, tempBytes float64
	for rows.Next() {
		if err := rows.Scan(&datname,
			&numbackends,
			&tupReturned,
			&tupFetched,
			&tupInserted,
			&tupUpdated,
			&tupDeleted,
			&xactCommit,
			&xactRollback,
			&blksRead,
			&blksHit,
			&conflicts,
			&deadlocks,
			&tempFiles,
			&tempBytes); err != nil {
			return err
		}

		// postgres_stat_database_numbackends
		ch <- prometheus.MustNewConstMetric(c.numbackends, prometheus.GaugeValue, numbackends, datname)
		// postgres_stat_database_tup_returned_total
		ch <- prometheus.MustNewConstMetric(c.tupReturned, prometheus.CounterValue, tupReturned, datname)
		// postgres_stat_database_tup_fetched_total
		ch <- prometheus.MustNewConstMetric(c.tupFetched, prometheus.CounterValue, tupFetched, datname)
		// postgres_stat_database_tup_inserted_total
		ch <- prometheus.MustNewConstMetric(c.tupInserted, prometheus.CounterValue, tupInserted, datname)
		// postgres_stat_database_tup_updated_total
		ch <- prometheus.MustNewConstMetric(c.tupUpdated, prometheus.CounterValue, tupUpdated, datname)
		// postgres_stat_database_tup_deleted_total
		ch <- prometheus.MustNewConstMetric(c.tupDeleted, prometheus.CounterValue, tupUpdated, datname)
		// postgres_stat_database_xact_commit_total
		ch <- prometheus.MustNewConstMetric(c.xactCommit, prometheus.CounterValue, xactCommit, datname)
		// postgres_stat_database_tup_xact_rollback_total
		ch <- prometheus.MustNewConstMetric(c.xactRollback, prometheus.CounterValue, xactRollback, datname)
		// postgres_stat_database_blks_read_total
		ch <- prometheus.MustNewConstMetric(c.blksRead, prometheus.CounterValue, blksRead, datname)
		// postgres_stat_database_blks_hit_total
		ch <- prometheus.MustNewConstMetric(c.blksHit, prometheus.CounterValue, blksHit, datname)
		// postgres_stat_database_conflicts_total
		ch <- prometheus.MustNewConstMetric(c.conflicts, prometheus.CounterValue, conflicts, datname)
		// postgres_stat_database_deadlocks_total
		ch <- prometheus.MustNewConstMetric(c.deadlocks, prometheus.CounterValue, deadlocks, datname)
		// postgres_stat_database_temp_files_total
		ch <- prometheus.MustNewConstMetric(c.tempFiles, prometheus.CounterValue, tempFiles, datname)
		// postgres_stat_database_temp_bytes_total
		ch <- prometheus.MustNewConstMetric(c.tempBytes, prometheus.CounterValue, tempBytes, datname)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	return nil
}
