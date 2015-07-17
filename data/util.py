from matplotlib.backends.backend_pdf import PdfPages
import matplotlib.pyplot as plt

def reset_axis():
    ax = plt.gca()  # get the current axes
    ax.relim()      # make sure all the data fits
    ax.autoscale()  # auto-scale

# lifted from stackoverflow
def multipage(filename, figs=None, dpi=200):
    pp = PdfPages(filename)
    if figs is None:
        figs = [plt.figure(n) for n in plt.get_fignums()]
    for fig in figs:
        fig.savefig(pp, format='pdf')
    pp.close()

# filters a dataframe, keeping rows with field in vals
# if vals is empty, user is prompted to enter some
def lock_float_field(df, field, vals=[]):
    if vals == []:
	vals = prompt(field)
	if vals == []:
	    return None

    filter = df[field].isin(vals)
    filtered = df[filter]
    return filtered

def prompt(field):
    vals = raw_input('make graphs where ' + field + " (space separated list)=")
    if vals == "":
        return []

    vals = vals.split()
    vals = map(float, vals)
    return vals
