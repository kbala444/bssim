import sqlite3
import numpy as np
import pandas as pd
import seaborn as sns
import os
import ConfigParser
import util
import matplotlib.pyplot as plt
import matplotlib.cm as cm
import sys

use_config = True

# array of tuples of all graphing functions in Grapher and a description for each
graph_funcs = []
# decorator to keep track of graphing functions
def is_graph(desc):
    # wrap add_func in is_graph to accept a description argument
    def add_func(func):
        graph_funcs.append((func, desc))
        return func

    return add_func

class Grapher():
    # create new grapher that reads data from given sqlite3 db and config file
    def __init__(self, db, cfg):
        self.conn = sqlite3.connect(db)
        self.cur = self.conn.cursor()

        # configure
        self.config = ConfigParser.ConfigParser()
        self.config.read(cfg)
        self.wl = self.config.get('general', 'workload')

        # load dataframe
        self.df = self.loaddf()

        self.lats = []
        self.bws = []
    
    # returns dataframe of runs with given workload
    def loaddf(self):
        print 'loading sql into dataframe...'
        #get mean block times where workload=samples/star order by runid ascneding
        cols_dict = {'runids': [], 'latencies' : [], 'bandwidths': [], 'durations': []}

        # assumes alphabetic order in select statement............
        sql = 'SELECT bandwidth, duration, latency, runid FROM runs where workload LIKE ("%" || ? || "%") ORDER BY runid ASC'
        for row in self.cur.execute(sql, (self.wl,)):
            i = 0
            for k in sorted(cols_dict.iterkeys()):
                cols_dict[k].append(row[i])
                i += 1

        df = pd.DataFrame.from_dict(cols_dict)

        return df
    
    # loads block_times into dataframe
    # do it per method so if no graphs involve mean block times, the runtime is much faster
    def load_block_times(self):
        if 'means' in self.df.columns:
            return

        means = []
        # get mean block time for each runid
        for runid in self.df['runids']:
            self.cur.execute('SELECT AVG(time) FROM block_times where runid=?', (runid,))
            means.append(self.cur.fetchone()[0])

        self.df['means'] = means

    @is_graph('graph of block times vs time for most recent run')
    def bttime(self):
        print 'loading block time vs time...'
        # get block_time rows for most recent run
        self.cur.execute('SELECT timestamp, time, runid FROM block_times where runid=(select max(runid) from runs)')
        rows = self.cur.fetchall()
        rid = (rows[0][2],)

        # get tuple reflecting run config to show under graph
        self.cur.execute('SELECT * FROM runs where runid=?', rid)
        config = self.cur.fetchone()
        config = map(str, config)
        names = [i[0] for i in self.cur.description]
        desc = str(zip(names, config))

        timestamps = []
        times = []
        for ts, time, rid in rows:
            timestamps.append(ts)
            times.append(time)
        
        timedf = pd.DataFrame.from_dict({'timestamps' : timestamps, 'times' : times})
        # change nanosecond timestamps to seconds
        timedf['timestamps'] = timedf['timestamps'].astype(float) / (1000 * 1000)
        g = sns.lmplot("timestamps", "times", data=timedf)
        print desc
        g.ax.set_title(self.wl)
        g.set_axis_labels("time (seconds)", "block times (ms)")
        # doesn't work...
        #g.ax.text(0.1, 0.1, desc)

    @is_graph('graph of latencies vs mean for given bandwidths')
    def latmean(self):
        print 'loading latency vs mean...'
        self.load_block_times()
        filtered = util.lock_float_field(self.df, 'bandwidths', self.bws)
        if filtered is None:
            return self.latmeanbw()

        g = sns.lmplot("latencies", "means", data=filtered[['latencies', 'means', 'bandwidths']], scatter=True, col='bandwidths') 
        g.set(ylim=(0, 200))

    @is_graph('graph of latencies vs block_times (colored by runid) for given bandwidths')
    def latmean_nodes(self):
        print 'loading latency vs mean all nodes displayed...'
        self.load_block_times()
        filtered = util.lock_float_field(self.df, 'bandwidths', self.bws)
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
        g = sns.lmplot("latencies", "times", data=timesdf[['latencies', 'times']],# 'bandwidths']], 
                scatter=True, scatter_kws={'c': timesdf['runids'].tolist(), 'cmap': cm.Accent, "alpha": .5}, legend_out=True)

    @is_graph('graph of latencies vs mean where bandwidth is the size of the point')
    def latmeanbw(self):
        # take log of bw array for better sizing
        self.load_block_times()
        normbws = np.array(self.df.bandwidths) 
        g = sns.lmplot("latencies", "means", data=self.df[['latencies', 'means']], scatter_kws={"s": np.log2(normbws) * 10, "alpha" : .5})
        g.set(ylim=(0, 400))
        g = self.with_title(g)

    @is_graph('graph of latencies vs simulation durations for given bandwidths')
    def latdur(self):
        print 'latency vs duration'
        filtered = util.lock_float_field(self.df, 'bandwidths', self.bws)
        if filtered is None:
            return self.latmeanbw()

        g = sns.lmplot("latencies", "durations", data=filtered[['latencies', 'durations', 'bandwidths']].astype(float), col='bandwidths')

    @is_graph('graph of bandwidths vs means for given latencies')
    def bwmeans(self):
        print 'bandwidth vs means'
        self.load_block_times()
        filtered = util.lock_float_field(self.df, 'latencies', self.lats)
        if filtered is None:
            return latmeanbw()
        
        filter = filtered["bandwidths"] > 0
        filtered = filtered[filter]

        # use plain pyplot cause seaborn has semilog issues
        plt.figure()
        plt.scatter(filtered["bandwidths"].tolist(), filtered["means"].tolist())
        plt.semilogx()
        plt.title(self.wl)
        plt.xlabel('bandwidth')
        plt.ylabel('duration')

    #@is_graph
    def bwdur(self):
        print 'bandwidth vs durations'
        filtered = util.lock_float_field(self.df, 'latencies', self.lats)
        if filtered is None:
            return latmeanbw()

        filter = filtered["bandwidths"] > 0
        filtered = filtered[filter]

        # use plain pyplot cause seaborn has semilog issues
        plt.figure()
        plt.scatter(filtered["bandwidths"].tolist(), filtered["means"].tolist())
        plt.semilogx()
        plt.title(self.wl)
        plt.xlabel('bandwidth')
        plt.ylabel('duration')

    @is_graph('graph of # of files completed for most recent run over time')
    def show_completion(self):
        self.cur.execute('SELECT * FROM file_times where runid=(select max(runid) from runs) order by timestamp asc')
        rows = self.cur.fetchall()

        timestamps = []
        for row in rows:
            # convert from microseconds to seconds
            timestamps.append(float(row[0]) / (1000 * 1000))

        counts = [i + 1 for i in xrange(len(rows))]
        
        plt.figure()
        plt.fill_between(timestamps, counts, 0)
        ax = plt.gca()
        ax.get_xaxis().get_major_formatter().set_useOffset(False)
        plt.xlabel("time (s)")
        plt.ylabel("received file count")

    @is_graph('graph of file size vs file times for given bandwidths')
    def fsize_time(self):
        # select rids where bandwidth in self.bws to create mapping of runids to bws
        runid_bw = {}
        self.cur.execute('SELECT runid, bandwidth FROM runs WHERE bandwidth IN (%s)' % ','.join('?'*len(self.bws)), self.bws)
        rows = self.cur.fetchall()
        for rid, bw in rows:
            runid_bw[rid] = bw

        # create dataframe of file times vs their size and bandwidth of that run
        runids = runid_bw.keys()
        df_dict = {'bandwidth': [], 'time': [], 'size': []}
        self.cur.execute('SELECT runid, time, size FROM file_times WHERE runid IN (%s)' % ','.join('?'*len(runids)), runids)
 
        rows = self.cur.fetchall()
        for runid, time, size in rows:
            df_dict['bandwidth'].append(runid_bw[runid])
            df_dict['time'].append(time)
            # convert size to kb
            df_dict['size'].append(float(size) / 1024)
        
        df = pd.DataFrame.from_dict(df_dict)
        g = sns.lmplot("size", "time", data=df, scatter=True, col='bandwidth') 
        g.set_axis_labels("file size (Kb)", "mean file time (s)")

    @is_graph('graph of strategy vs files times for manual links')
    def strategy_times(self):
        # split on workload, boxplot based on strategy and duration stats, 
        df_dict = {'time': [], 'workload': [], 'strategy': []}

        self.cur.execute('SELECT runid, workload, duration, strategy FROM runs WHERE manual=1')
        rows = self.cur.fetchall()
        for rid, wl, d, strat in rows:
            #df['id'].append(rid)
            df_dict['workload'].append(wl)
            df_dict['strategy'].append(strat)
            df_dict['time'].append(d)

        df = pd.DataFrame.from_dict(df_dict)
        g = sns.factorplot(x='strategy', y='time', data=df, col='workload', kind='box')

    # saves/shows graphs if specified in config and closes connection
    def finish(self):
        if self.config.getboolean('general', 'save'):
            util.multipage(self.config.get('general', 'filename'))

        if self.config.getboolean('general', 'show'):
            sns.plt.show()

        self.conn.close()

    def with_title(self, g):
        for axes in g.axes:
            for ax in axes:
                ax.set_title(self.wl)
        return g

   
# determines what graphs to show user with prompt
def pick_figs(grapher):
    print 'Which graphs would you like (space separated list):'
    i = 0
    for f in graph_funcs:
        print '%d: %s' % (i, f[1])
        i += 1

    inp = raw_input('\n->')
    figs = inp.split(' ')
    # run them all by default
    if inp == '':
        figs = [i for i in xrange(len(graph_funcs))]

    figs = map(int, figs)
    return figs

# gets figs from config file
def read_figs(grapher):
    cfg = grapher.config
    figs = cfg.get('graphs', 'graphs')
    figs = figs.split()
    figs = map(int, figs)

    lats = cfg.get('graphs', 'latencies')
    grapher.lats = map(int, lats.split())

    bws = cfg.get('graphs', 'bandwidths')
    grapher.bws = map(int, bws.split())

    return figs

def main():
    grapher = Grapher('metrics', 'config.ini')
    if len(sys.argv) > 1 and sys.argv[1] == '-p':
        figs = pick_figs(grapher)
    else:
        figs = read_figs(grapher)

    for index in figs:
        graph_funcs[index][0](grapher)

    grapher.finish()

if __name__ == "__main__":
    main()
    
