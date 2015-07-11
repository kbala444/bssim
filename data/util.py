from matplotlib.backends.backend_pdf import PdfPages
import matplotlib.pyplot as plt

def reset_axis():
    ax = plt.gca()  # get the current axes
    ax.relim()      # make sure all the data fits
    ax.autoscale()  # auto-scale

def multipage(filename, figs=None, dpi=200):
    pp = PdfPages(filename)
    if figs is None:
        figs = [plt.figure(n) for n in plt.get_fignums()]
    for fig in figs:
        fig.savefig(pp, format='pdf')
    pp.close()

def lock_float_field(df, field):
    val = raw_input(field + "=")
    if val == "":
        return None

    val = int(val)
    filter = df[field] == val
    filtered = df[filter]
    return filtered

