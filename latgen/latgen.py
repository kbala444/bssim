from lxml import html
import numpy as np
import requests
import random
import sys
import networkx as nx
from matplotlib import pyplot as plt
#import seaborn as sns
import re
import math
from optparse import OptionParser

class NetworkGenerator():
    # inits a new NetworkGenerator with n nodes
    # cities and latencies from http://ipnetwork.bgtmo.ip.att.net/pws/network_delay.html
    def __init__(self, n):
        # saved subset of above att link
        self.n = n
        tree = html.parse('ogpage.html')
        self.cities = self.get_cities(tree)
        self.lats = self.get_latencies(tree)
        # list of cities representing node locations
        self.nodes = self.gen_nodes()

    def gen_nodes(self):
        nodes = []
        for node in xrange(self.n):
            nodes.append(random.choice(self.cities))

        return nodes

    # returns link options for n interconnected peers 
    # bandwidth is sampled from a normal distribution with mean mean_bw
    def gen_connections(self, mean_bw):
        #http://www.netindex.com/download/2,1/United-States/
        randoms = np.random.normal(mean_bw, 10, self.n)

        locs = self.nodes
        
        # why are links 2 way? why 2 links between each pair?
        # how to allocate bw between multiple connections?
        # (n1, n2, lat, bw)
        cons = []
        for i in xrange(len(locs)):
            for j in xrange(i + 1, len(locs)):
                l = self.get_lat_btwn(locs[i], locs[j])
                cons.append((i, j, l, randoms[i]))

        return cons
    
    # formats conn tuples into 'n1->n2 lat bw' format
    def format_cons(self, cons):
        out = []
        for con in cons:
            out.append('%d->%d %d %d' % (con[0], con[1], con[2], con[3]))

        return out
            
    # returns list of cities from that att site
    def get_cities(self, tree):
        cities = []
        city_rows = tree.xpath("//td[@nowrap]")
        for row in city_rows:
            name = row.text_content().strip().lower()
            if len(name) > 3:
                cities.append(name)

        return cities

    # makes those green ATT boxes into 2d array
    def get_latencies(self, tree):
        lat_array = [[] for i in xrange(24)]
        vals = tree.xpath("//td[@bgcolor='#66CC66']")

        cap = 1
        pos = 0
        while pos < len(vals):
            for i in xrange(cap):
                lat_array[cap - 1].append(vals[pos])
                pos += 1

            cap += 1

        return lat_array

    # gets latency between two cities, c1 and c2
    def get_lat_btwn(self, c1, c2):
        if c1 not in self.cities or c2 not in self.cities:
            raise ValueError('Cities must be in cities array')

        # indexes of cities in city array, used to lookup latencies
        i1 = self.cities.index(c1)
        i2 = self.cities.index(c2)

        # must look at how att website looks for this to make sense
        if i1 > i2:
            # add 1 to indexes because i skipped the atlanta column
            res = self.lats[i1][i2 + 1].text_content()
        elif i1 == i2:
            # what to do if same city?  maybe sample from smaller latencies?
            # i'll just return 2 for now lol
            res = 2
        else:
            res = self.lats[i2][i1 + 1].text_content()

        return int(str(res))

    # gets population information for each city in self.cities
    # i made this and then realized its not useful but maybe i'll figure out a use for it
    def get_pops(self):
        print 'fetching population information...'
        pops = []
        req = "http://api.wolframalpha.com/v2/query?appid=586XQG-QXXJ4K6244&input=population%%20of%%20%s&format=plaintext"
        for city in lg.cities:
            r = requests.get(req % city)
            tree = html.fromstring(r.content)
            res = tree.xpath('//plaintext')
            pop = res[1].text_content().partition(' ')[0]
            pops.append(pop)

        return pops
    
    # returns map of cities to long-lat from local file
    def get_locs(self):
        f = open('longlat', 'r')
        locs = {}
        for line in f.readlines():
            line = line.split()
            # join names with spaces in them
            line[0:-2] = [' '.join(line[0:-2])]
            locs[line[0]] = [float(line[1]), float(line[2])]

        return locs
        
    def graph_network(self, cons):
        g = nx.Graph()
        #g.add_nodes_from(self.nodes)

        for con in cons:
            g.add_edge(con[0], con[1], weight=con[2])

        #nx.draw(g)
        # use pos to show node location?
        pos = {}
        labels = {}
        locs = self.get_locs()
        used_cities = []
        for node in xrange(len(self.nodes)):
            city = self.nodes[node]
            # could add random noise to every lat long as well
            loc = locs[self.nodes[node]]
            if city in used_cities:
                loc[0] += 5
            else:
                used_cities.append(city)

            pos[node] = loc
            labels[node] = '%d %s' % (node, city)


        #pos = nx.circular_layout(g)
        edge_labels = {}
        edges = g.edges()
        for i in xrange(len(g.edges())):
            edge_labels[edges[i]] = '%d, %d' % (cons[i][2], cons[i][3]) 

        nx.draw(g, pos, node_size=1000, alpha=.5)
        #nx.draw_networkx_edge_labels(g, pos, edge_labels, font_size=14)
        nx.draw_networkx_labels(g, pos, labels, font_size=12, label_pos=0)
        print g.number_of_nodes()
        plt.draw()

# puts lat and long in latlong file using geoip 
def update_locs():
    from geopy.geocoders import Nominatim

    lg = NetworkGenerator(10)
    
    f = open('longlat', 'w')
    geolocator = Nominatim()
    locs = []
    for city in lg.cities:
        loc = geolocator.geocode({'city': city})
        print loc.address
        f.write('%s %.5f %.5f\n' % (city, loc.longitude, loc.latitude))
       
def insert_into_wl(filepath, cons):
    # read original wl
    f = open(filepath, 'r+')
    old = f.readlines()
    
    # overwrite with conns added
    old.insert(1, '\n'.join(cons) + '\n')
    f.seek(0)
    f.truncate()
    f.write(''.join(old))
    f.close()

def main():
    parser = OptionParser()
    parser.add_option("-f", "--file", dest="filename",
                              help="write report to FILE", metavar="FILE")
    parser.add_option("-f", "--file", action="store", type="string", dest="filename")

    lg = NetworkGenerator(10)
    cons = lg.gen_connections(37.5)
    lg.graph_network(cons)

    formatted = lg.format_cons(cons)

    if len(sys.argv) > 1 and '-i' in sys.argv:
        findex = sys.argv.index('-i') + 1
        insert_into_wl(sys.argv[findex], formatted)
    else:
        print '\n'.join(formatted)

    plt.show()


# add opt to show hide lat bw
if __name__ == "__main__":
    # update if -u flag
    if len(sys.argv) > 1 and '-u' in sys.argv:
        print 'updating locs'
        update_locs()
        print 'done'

    main()

