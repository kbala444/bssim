import sqlite3
import numpy as np
import pandas as pd
import seaborn as sns
import os
import ConfigParser
import util

class Grapher():
    # create new grapher that reads data from given sqlite3 db and config file
    def __init__(self, db, cfg):
        self.conn = sqlite3.connect('metrics')
        self.cur = self.conn.cursor()

        # configure
        self.config = ConfigParser.ConfigParser()
        self.config.read('config.ini')
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
        g = sns.lmplot("bandwidths", "means", data=filtered[['bandwidths', 'means', 'latencies']], col='latencies')

    def bwdur(self):
        print 'bandwidth vs durations'
        filtered = util.lock_float_field(self.df, 'latencies')
        if filtered is None:
            return latmeanbw()

        filter = filtered["bandwidths"] > 0
        filtered = filtered[filter]
        g = sns.lmplot("bandwidths", "durations", data=filtered[['bandwidths', 'durations', 'latencies']], col='latencies')

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
    grapher = Grapher('metrics', 'config.ini')

    print 'Which graphs would you like (space separated list):'
    print '0: time series of block times'
    print '1: latency v means'
    print '2: latency v duration'
    print '3: bandwidths v means'
    print '4: bandwidths v durations'
    print '5: latency v means v bw (size of circle is bw)'

    funcs = [grapher.bttime, grapher.latmean, grapher.latdur, 
            grapher.bwmeans, grapher.bwdur, grapher.latmeanbw]

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
