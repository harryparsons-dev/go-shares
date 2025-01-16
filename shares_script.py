from collections import defaultdict
import pandas as pd
import weasyprint
import webbrowser
import datetime
import os
import yfinance as yf
import math
import matplotlib.pyplot as plt
import sys


def highlight_values(x):

    if isinstance(x, (int, float)):
        if x < 0:
            return 'color: red'
        elif x >= 0:
            return 'color: black'
        else:
            pass



def getData(ticker, key):
    try:
        x = str(ticker)
        sym = ""
        if (x[-1:]) == ".":
            sym += x + "L"
        else:
            sym += x + ".L"
        stock = yf.Ticker(sym)
        stock= stock.info

        return stock[key]

    except:
        pass


def getSector(ticker):
    x = getData(ticker, 'sector')
    return x
def getDividendYield(ticker):
    x =getData(ticker, 'dividendYield')
    if not x ==  None:
        return ("{:.2f}".format(x*100))
    else:
        pass




# for csv in files:



df = pd.read_csv(sys.argv[1])
pd.set_option('display.width', 300)




df.columns = df.columns.str.strip('\ufeff')
df['Qty'] = df['Qty'].astype('Int64').astype('string')

df = df.drop(columns=['Day Gain/Loss', 'Day Gain/Loss %', 'Market Value £'])
df["Gain/Loss %"] = df["Gain/Loss %"].str.rstrip("%").astype('float')

totals = df.tail(2)
df.drop(df.tail(2).index,inplace=True)







# %%
df['Sector'] = df['Symbol'].apply(getSector)

# %%
df["Dividend Yield"] = df['Symbol'].apply(getDividendYield)

# %%
df = df.sort_values(by="Gain/Loss %", ascending=False)

# %%
result = pd.concat([df,totals])
# result = result.fillna('')



# %%
result['Market Value'] = result['Market Value'].str.replace('£', '')
result['Market Value'] = result['Market Value'].str.replace(',', '')
result['Market Value'] = result['Market Value'].astype('float')

# %%
result['Gain/Loss']= result["Gain/Loss"].str.replace('£', '')
result['Gain/Loss']= pd.to_numeric(result["Gain/Loss"].str.replace(',', '') ,errors='coerce')



# %%
result['Gain/Loss %']= pd.to_numeric(result["Gain/Loss %"],errors='coerce')

# %%
result = result.reset_index(drop=True)
result = result.fillna('')

# %%
result['Gain/Loss'] = pd.to_numeric(result['Gain/Loss'], errors='coerce')
result['Gain/Loss %'] = pd.to_numeric(result['Gain/Loss %'], errors='coerce')
result['Market Value'] = pd.to_numeric(result['Market Value'], errors='coerce')

# %%
# Apply styling with proper formatting
def custom_format(x, column_name=''):
    if isinstance(x, str):
        return x
    elif pd.isna(x):  # Check if the value is NaN
        return ''  # Return an empty string for NaN
    elif isinstance(x, (int, float)):  # Check if it's a number
        # Apply custom formatting for 'Gain/Loss' and 'Market Value'
        if column_name == 'Gain/Loss' or column_name == 'Market Value':
            return f'£{x:,.2f}'  # Format with comma separator and 2 decimal places
        elif column_name == 'Gain/Loss %':
            return f'{x:.2f}%'  # Format as percentage
    return x  # Default return for other types

# Apply styling with proper formatting using a conditional function
styled = result.style.format({
    'Gain/Loss': lambda x: custom_format(x, 'Gain/Loss'),
    'Gain/Loss %': lambda x: custom_format(x, 'Gain/Loss %'),
    'Market Value': lambda x: custom_format(x, 'Market Value'),
    }).applymap(highlight_values, subset=['Gain/Loss', 'Gain/Loss %'])

# %%
newdf = df
newdf['Market Value'] = newdf['Market Value'].str.replace('£', '')
newdf['Market Value'] = newdf['Market Value'].str.replace(',', '')
newdf['Market Value'] = newdf['Market Value'].astype('float')

# %%

sector_market_value = newdf.groupby('Sector')['Market Value'].sum()

# %%
sector_market_value

# %% [markdown]
# 

# %%
plt.figure(figsize=(8, 6))
def custom_autopct(pct):
    return ('%.1f%%' % pct) if pct >= 4 else ''
wedges, texts, autotexts = plt.pie(sector_market_value, autopct=custom_autopct, startangle=90)

total = sum(sector_market_value)
percentages = [value / total * 100 for value in sector_market_value]

# Create custom labels for the legend
legend_labels = []

# Create legend labels using a for loop
for i in range(len(sector_market_value.index)):
    name = sector_market_value.index[i]
    pct = percentages[i]
    if pct <= 4:
        legend_labels.append(f'{name}: {pct:.2f}%')
    else:
         legend_labels.append(f'{name}')


# For percentages below 4%, set the labels manually to be outside the pie chart
for i, autotext in enumerate(autotexts):
    pct_value = sector_market_value[i] / sum(sector_market_value) * 100
    if pct_value < 4:
        # Move the label slightly further outside the pie slice
        autotext.set_position((1.1, autotext.get_position()[1]))

plt.title('Market Value Proportion by Sector')

plt.legend(wedges, legend_labels, title="Sectors", loc="center left", bbox_to_anchor=(1, 0, 0.5, 1))

plt.axis('equal')

plt.savefig('market_value_pie_chart.pdf', format='pdf', bbox_inches='tight')
# ')  # Equal aspect ratio ensures the pie chart is circular
plt.show()

# %%
HTML_TEMPLATE = '''
<html>
<head>
<style>
@page {
    size: landscape;
    margin: 0in 0in 0in 0in;
}
body {
    margin: 0;
}
div {
    position: relative;
    font-size: 8px;
    width: 10%%;
    font-family: "Courier New";
    padding-top: 5px;
    font-weight: bold;
}
table {
    margin-left: 0;
    margin-right: 0;
    margin-top: 0;
    font-size: %(font_size)dpx;
}
table, th, td {
    border: 0.2px solid black;
    border-collapse: collapse;
}
th, td {
    padding: %(padding)dpx;
    text-align: left;
    font-family: Helvetica, Arial, sans-serif;
}
table tbody tr:hover {
    background-color: #dddddd;
}
table thead th {
    text-align: center;
}
.wide {
    width: 90%%;
}
</style>
</head>
<body>
'''

# %%
HTML_TEMPLATE2 = '''
</body>
</html>
'''



# with open('small-template.html', 'r', encoding='utf-8') as smallTemplate:
#     small_template = smallTemplate.read()



def to_html_pretty(df, html_template,filename='out.html', title=''):


    options = {
        "font_size" : float(sys.argv[2]),
        "padding" : float(sys.argv[3]),
    }

    try:
        styled_template = html_template % options
    except KeyError as e:
        print(f"KeyError: {e}. Check your template placeholders or options dictionary.")
        return

    ht = ''

    ht += df.to_html(classes='wide', escape=False, index=False)
    if title != '':
        ht += '<div> %s </div>\n' % title

    with open(filename, 'w', encoding='utf-8') as f:
        f.write(styled_template+ ht + HTML_TEMPLATE2)

# Pretty print the dataframe as an html table to a file
intermediate_html = 'intermediate.html'

x = datetime.datetime.now()
title = 'Updated: ' + x.strftime('%b') + ' ' + x.strftime('%y')

# Normal html:
to_html_pretty(styled,HTML_TEMPLATE,intermediate_html, title )
# Bob template:

# if you do not want pretty printing, just use pandas:
# result.to_html(intermediate_html)

# Create the full path including directories if they don't exist
filepath = sys.argv[4]
output_dir = os.path.dirname(filepath)
if output_dir:  
    os.makedirs(output_dir, exist_ok=True)


out_pdf = filepath + '.pdf'


weasyprint.HTML('intermediate.html').write_pdf(out_pdf)




# to_html_pretty(styled,bob_template,intermediate_html, title )
# weasyprint.HTML('intermediate.html').write_pdf('bob.pdf')
#  webbrowser.open_new_tab("/tmp/intermediate.html")


# This is the table pretty printer used above:


