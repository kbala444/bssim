import sqlite3
import numpy as np
import pandas as pd
import seaborn as sns
import os
import ConfigParser
import util
import matplotlib.pyplot as plt
import matplotlib.cm as cm

class Grapher():
    # create new grapher that reads data from given sqlite3 db and config file
    def __init__(self, db, cfg):
        self.conn = sqlite3.connect('metrics')
        self.cur = self.conn.cursor()

        # configure
        self.config = ConfigParser.ConfigParser()
        self.config.read(cfg)
        self.wl = self.config.get('DEFAULT', 'workload')

        # load dataframe
        self.df = self.loaddf(self.wl)

    # returns dataframe of runs with given workload
    def loaddf(self, workload):
        print 'loading sql into dataframe...'
        #get mean block times where workload=samples/star order by runid ascneding
        cols_dict = {'runids': [], 'latencies' : [], 'bandwidths': [], 'durations': []}

        # assumes alphabetic order in select statement............
        sql = 'SELECT bandwidth, duration, latency, runid FROM runs where workload LIKE ("%" || ? || "%") ORDER BY runid ASC'
        for row in self.cur.execute(sql, (workload,)):
            i = 0
            for k in sorted(cols_dict.iterkeys()):
                cols_dict[k].append(row[i])
                i += 1

        means = []
        # get mean block time for each runid
        for runid in cols_dict['runids']:
            self.cur.execute('SELECT AVG(time) FROM block_times where runid=?', (runid,))
            means.append(self.cur.fetchone()[0])

        cols_dict['means'] = means
        df = pd.DataFrame.from_dict(cols_dict)

        # drop tables without durations
        # make sure to clean df before any plots cause mean outliers are associated with no durations
        df = df[df['durations'].astype(object) != ""]
        
        return df

    # graph of block times over time for most recent run
    def bttime(self):
        # get block_time rows for most recent run, assumes most recent run isn't in progress
        self.cur.execute('SELECT * FROM block_times where runid=(select max(runid) from runs)')
        rows = self.cur.fetchall()
        rid = (rows[0][2],)

        self.cur.execute('SELECT * FROM runs where runid=?', rid)
        config = self.cur.fetchone()
        config = map(str, config)
        names = [i[0] for i in self.cur.description]

        timestamps = []
        times = []
        for row in rows:
            timestamps.append(row[0])
            times.append(row[1])
        
        timedf = pd.DataFrame.from_dict({'timestamps' : timestamps, 'times' : times})
        g = sns.lmplot("timestamps", "times", data=timedf)
        g.ax.text(0.005, 0.005, str(zip(names, config)))
        g.ax.set_title(self.wl)

    # graphs latencies vs mean for given bandwidth
    def latmean(self):
        print 'latency vs mean'
        filtered = util.lock_float_field(self.df, 'bandwidths')
        if filtered is None:
            return self.latmeanbw()

        g = sns.lmplot("latencies", "means", data=filtered[['latencies', 'means', 'bandwidths']], scatter=True, col='bandwidths') 
                #scatter_kws={'c':filtered['runids'].tolist(), 'cmap': cm.Accent})

    def latmean_nodes(self):
        print 'latency vs mean all nodes displayed'
        filtered = util.lock_float_field(self.df, 'bandwidths')
        if filtered is None:
            return self.latmeanbw()

        all_times_dict = {'runids': [], 'latencies': [], 'bandwidths': [], 'times': []}
        for runid in filtered['runids']:
            # get latency for runid
            self.cur.execute('SELECT latency, bandwidth FROM runs where runid=?', (runid,))
            lat, bw = self.cur.fetchone()

            # get block times from runid and populate bandwidths and latencies
            for row in self.cur.execute('SELECT time FROM block_times where runid=?', (runid,)):
                all_times_dict['runids'].append(runid)
                all_times_dict['latencies'].append(lat)
                all_times_dict['bandwidths'].append(bw)
                all_times_dict['times'].append(row[0])

        timesdf = pd.DataFrame.from_dict(all_times_dict)
        print timesdf['runids']
        g = sns.lmplot("latencies", "times", data=timesdf[['latencies', 'times']],# 'bandwidths']], 
                scatter=True, scatter_kws={'c': timesdf['runids'].tolist(), 'cmap': cm.Accent, "alpha": .5}, legend_out=True)

    def latmeanbw(self):
        # take log of bw array for better sizing
        normbws = np.array(self.df.bandwidths) 
        g = sns.lmplot("latencies", "means", data=self.df[['latencies', 'means']], scatter_kws={"s": np.log2(normbws) * 10, "alpha" : .5})
        g.set(ylim=(0, 400))
        g = self.with_title(g)

    def latdur(self):
        print 'latency vs duration'
        filtered = util.lock_float_field(self.df, 'bandwidths')
        if filtered is None:
            return self.latmeanbw()

        g = sns.lmplot("latencies", "durations", data=filtered[['latencies', 'durations', 'bandwidths']].astype(float), col='bandwidths')

    def bwmeans(self):
        print 'bandwidth vs means'
        filtered = util.lock_float_field(self.df, 'latencies')
        if filtered is None:
            return latmeanbw()
        
        filter = filtered["bandwidths"] > 0
        filtered = filtered[filter]

        # use plain pyplot cause seaborn has semilog issues
        plt.scatter(filtered["bandwidths"].tolist(), filtered["means"].tolist())
        plt.semilogx()
        plt.title(self.wl)
        plt.xlabel('bandwidth')
        plt.ylabel('duration')

    def bwdur(self):
        print 'bandwidth vs durations'
        filtered = util.lock_float_field(self.df, 'latencies')
        if filtered is None:
            return latmeanbw()

        filter = filtered["bandwidths"] > 0
        filtered = filtered[filter]

        # use plain pyplot cause seaborn has semilog issues
        plt.scatter(filtered["bandwidths"].tolist(), filtered["means"].tolist())
        plt.semilogx()
        plt.title(self.wl)
        plt.xlabel('bandwidth')
        plt.ylabel('duration')

    def show_completion(self):
        self.cur.execute('SELECT * FROM file_times where runid=(select max(runid) from runs) order by timestamp asc')
        rows = self.cur.fetchall()

        timestamps = []
        for row in rows:
            timestamps.append(row[0])

        counts = [i + 1 for i in xrange(len(rows))]

        plt.fill_between(timestamps, counts, 0)
        plt.xlabel("time")
        plt.ylabel("received file count")

    # saves/shows graphs if specified in config and closes connection
    def finish(self):
        if self.config.getboolean('DEFAULT', 'save'):
            util.multipage(self.config.get('DEFAULT', 'filename'))

        if self.config.getboolean('DEFAULT', 'show'):
            sns.plt.show()

        self.conn.close()

    def with_title(self, g):
        for axes in g.axes:
            for ax in axes:
                ax.set_title(self.wl)
        return g

if __name__ == "__main__":
    # should get config file path from bash script
    grapher = Grapher('metrics', 'config.ini')

    print 'Which graphs would you like (space separated list):'
    print '0: time series of block times'
    print '1: latency v means'
    print '2: latency v duration'
    print '3: bandwidths v means'
    print '4: bandwidths v durations'
    print '5: latency v means v bw (size of circle is bw)'
    print '6: cummulative file completion times'
    print '7: latency v means but show all nodes (note: takes a couple seconds)'

    funcs = [grapher.bttime, grapher.latmean, grapher.latdur, 
            grapher.bwmeans, grapher.bwdur, grapher.latmeanbw, grapher.show_completion,
            grapher.latmean_nodes]

    inp = raw_input('\n->')
    figs = inp.split(' ')
    # run them all by default
    if inp == '':
        figs = [i for i in range(len(funcs))]

    figs = map(int, figs)

    i = 0
    for f in figs:
        funcs[f]()
        i += 1

    grapher.finish()
