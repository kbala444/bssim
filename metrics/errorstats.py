import pandas as pd
# map of workload to list of info tuples
data = {}
headers = ['mean block times', 'blocks received', 'duplicate blocks']

with open('data', 'r') as f:
    curr_workload = ''
    for line in f.readlines():
        if line[0] == '#' or line.isspace():
            continue

        if "/" in line:
            line = line.strip()
            data[line] = []
            curr_workload = line
        else:
            split = [x.strip() for x in line.split(',')[1:]]
            split[0] = split[0][:-2]
            split = map(float, split)
            data[curr_workload].append(tuple(split))

#info tuple has the form (mbt, blocks received, dup)
for workload in data:
    print 'workload', workload
    df = pd.DataFrame(data[workload], columns=headers)
    print 'means'
    print df.mean()
    print '\n'

    print 'standard deviations'
    print df.std()
    print '\n'

    print 'ranges'
    print (df.max() - df.min())
    print '\n\n\n'
