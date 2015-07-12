import sqlite3
import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
import matplotlib.dates as mdates
import seaborn as sns
import os
import ConfigParser
import util

def loaddf(workload):
    # get mean block times where workload=samples/star order by runid ascneding
    cols_dict = {'runids': [], 'latencies' : [], 'bandwidths': [], 'durations': []}

    # assumes alphabetic order in select statement............
    sql = 'SELECT bandwidth, duration, latency, runid FROM runs where workload LIKE ("%" || ? || "%") ORDER BY runid ASC'
    for row in c.execute(sql, (workload,)):
        i = 0
        for k in sorted(cols_dict.iterkeys()):
            cols_dict[k].append(row[i])
            i += 1

    means = []
    # get mean block time for each runid
    for runid in cols_dict['runids']:
        c.execute('SELECT AVG(time) FROM block_times where runid=?', (runid,))
        means.append(c.fetchone()[0])

    cols_dict['means'] = means
    df = pd.DataFrame.from_dict(cols_dict)

    # drop tables without durations
    # make sure to clean df before any plots cause mean outliers are associated with no durations
    df = df[df['durations'].astype(object) != ""]
    return df

# graph of block times over time for most recent run
def bttime():
    util.reset_axis()
    #ax = plt.gca()
    #ax.set_xlabel('time')
    #ax.set_ylabel('block time')
    # get block_time rows for most recent run, assumes most recent run isn't in progress
    c.execute('SELECT * FROM block_times where runid=(select max(runid) from runs)')
    rows = c.fetchall()
    rid = (rows[0][2],)

    c.execute('SELECT * FROM runs where runid=?', rid)
    config = c.fetchone()
    config = map(str, config)
    names = [i[0] for i in c.description]

    timestamps = []
    times = []
    for row in rows:
        timestamps.append(row[0])
        times.append(row[1])
    
    timedf = pd.DataFrame.from_dict({'timestamps' : timestamps, 'times' : times})
    fg = sns.lmplot("timestamps", "times", data=timedf)
    fg.ax.text(0.005, 0.005, str(zip(names, config)))

# graphs latencies vs mean for given bandwidth
def latmean():
    print 'latency vs mean'
    filtered = util.lock_float_field(df, 'bandwidths')
    if filtered is None:
        return latmeanbw()

    sns.lmplot("latencies", "means", data=filtered[['latencies', 'means']], scatter=True)

def latmeanbw():
    # take log of bw array for better sizing
    normbws = np.array(df.bandwidths) 
    sns.lmplot("latencies", "means", data=df[['latencies', 'means']], scatter_kws={"s": np.log(normbws) * 10, "alpha" : .5})

def latdur():
    print 'latency vs duration'
    filtered = util.lock_float_field(df, 'bandwidths')
    if filtered is None:
        return latmeanbw()

    sns.lmplot("latencies", "durations", data=df[['latencies', 'durations']].astype(float))

def bwmeans():
    print 'bandwidth vs means'
    filtered = util.lock_float_field(df, 'latencies')
    if filtered is None:
        return latmeanbw()
    
    filter = filtered["bandwidths"] > 0
    filtered = filtered[filter]
    grid = sns.lmplot("bandwidths", "means", data=filtered[['bandwidths', 'means']])

def bwdur():
    print 'bandwidth vs durations'
    filtered = util.lock_float_field(df, 'latencies')
    if filtered is None:
        return latmeanbw()

    filter = filtered["bandwidths"] > 0
    filtered = filtered[filter]
    sns.lmplot("bandwidths", "durations", data=filtered[['bandwidths', 'durations']])

conn = sqlite3.connect('metrics')
c = conn.cursor()

if __name__ == "__main__":
    config = ConfigParser.ConfigParser()
    config.read('config.ini')
    wl = config.get('DEFAULT', 'workload')

    df = loaddf(wl)

    print 'Which graphs would you like:'
    print '0: time series of block times'
    print '1: latency v means'
    print '2: latency v duration'
    print '3: bandwidths v means'
    print '4: bandwidths v durations'
    print '5: latency v means v bw (size of circle is bw)'

    funcs = [bttime, latmean, latdur, bwmeans, bwdur, latmeanbw]

    inp = raw_input('\n->')
    figs = inp.split(' ')
    # run them all by default
    if inp == '':
        figs = [i for i in range(len(funcs))]

    figs = map(int, figs)

    print df.dtypes

    plt.close('all')
    sns.lmplot("latencies", "means", data=df, row="bandwidths")
    sns.plt.show()

    i = 0
    # blank graph problem has something to do with figure here i think...
    for f in figs:
        #plt.figure(i)
        funcs[f]()
        util.reset_axis()
        i += 1

    if config.getboolean('DEFAULT', 'save'):
        util.multipage(config.get('DEFAULT', 'filename'))

    if config.getboolean('DEFAULT', 'show'):
        sns.plt.show()

    conn.close()
