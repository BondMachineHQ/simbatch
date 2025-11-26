#!/usr/bin/env python3

import sys
import subprocess
import getopt

def usage():
	print("SimBatch: A batch simulator for BondMachine designs")
	print("")
	print("SimBatch allows you to run batch simulations of BondMachine designs using a CSV input file and producing a CSV output file.")
	print("It expects the BondMachine design to be already compiled in the working directory called `bondmachine.json`.")
	print("")
	print("Usage: simbatch [options]")
	print("Options:")
	print("  -w, --working-dir DIR        set the working directory (default: working_dir)")
	print("  -i, --input-file FILE        set the input CSV file (default: simbatch_input.csv)")
	print("  -o, --output-file FILE       set the output CSV file (default: working_dir/simbatch_output.csv)")
	print("  -s, --simulation-steps N     number of simulation steps (default: 200)")
	print("  -m, --ml                     enable ML output formatting (probabilities + classification)")
	print("  -b, --benchcore              enable benchcore mode")
	print("  -H, --header                 include header row in output CSV")
	print("  -P, --prefix                 include data type prefix in output CSV")
	print("  -d, --data-type TYPE         data type for outputs (e.g. float32) (default: float32)")
	print("  --linear-data-range RANGE    pass a linear data range option to bondmachine/bmnumbers")
	print("  -v, --stop-on-valid-of N     stop on valid of output index N")
	print("  -h, --help                   show this help message and exit")
	print("")
	print("Example:")
	print("  simbatch.py -w working_dir -i input.csv -o out.csv -s 200")

working_dir=""
input_file=""
output_file=""
simulation_steps="200"
isml=False
benchCore=False
stopOnValidOf=-1
linear_data_range=""
data_type="float32"
prefix="0f"
header=False
omit_prefix=True

# Parse command line options
try:
	opts, args = getopt.getopt(sys.argv[1:], "w:i:o:s:mbd:v:hH", ["working-dir=","input-file=","output-file=", "simulation-steps=","ml", "benchcore", "linear-data-range=","data-type=", "stop-on-valid-of=", "help", "header"])
except  getopt.GetoptError:
	usage()
	sys.exit(2)


for o, a in opts:
	if o in ("-w", "--working-dir"):
		working_dir = a
	elif o in ("-i", "--input-file"):
		input_file = a
	elif o in ("-o", "--output-file"):
		output_file = a
	elif o in ("-s", "--simulation-steps"):
		simulation_steps = a
	elif o in ("--linear-data-range"):
		linear_data_range = "-linear-data-range "+a
	elif o in ("-d", "--data-type"):
		data_type = a
	elif o in ("-m", "--ml"):
		isml = True
	elif o in ("-b", "--benchcore"):
		benchCore = True
	elif o in ("-h", "--help"):
		usage()
		sys.exit(0)
	elif o in ("-v", "--stop-on-valid-of"):
		stopOnValidOf = int(a)
	elif o in ("-H", "--header"):
		header = True
	elif o in ("-P", "--prefix"):
		omit_prefix = False

if working_dir == "":
	working_dir="working_dir"
if input_file == "":
	input_file="simbatch_input.csv"
if output_file == "":
	output_file=working_dir+"/simbatch_output.csv"

last_step = str(int(simulation_steps) - 1)

# Load the BondMachine inputs
command="bondmachine -bondmachine-file "+working_dir+"/bondmachine.json -list-inputs "+ linear_data_range
p = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, close_fds=True)
p.wait()
inputs={}
if p.returncode==0:
	while True:
		o = p.stdout.readline().decode()
		if o == '' and p.poll() != None:
			break
		result=o.split()
		if len(result)==2:
			inputs[result[0]]=result[1]

# Load the BondMachine outputs
command="bondmachine -bondmachine-file "+working_dir+"/bondmachine.json -list-outputs "+linear_data_range
p = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, close_fds=True)
p.wait()
outputs={}
if p.returncode==0:
	while True:
		o = p.stdout.readline().decode()
		if o == '' and p.poll() != None:
			break
		result=o.split()
		if len(result)==2:
			outputs[result[0]]=result[1]

# Get the number prefix
command="bmnumbers -get-prefix " + data_type +" "+linear_data_range
p = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, close_fds=True)
p.wait()
if p.returncode==0:
	while True:
		o = p.stdout.readline().decode()
		if o == '' and p.poll() != None:
			break
		prefix=o.strip()

# print (prefix)
# print (inputs)
# print (outputs)

# Open the input file
input_file_handle=open(input_file, "r")
output_file_handle=open(output_file, "w")

# Write the output file header if needed
if header:
	if isml:
		for i in range(len(outputs)):
			output_file_handle.write("probability_"+str(i)+",")
		output_file_handle.write("classification")
	if benchCore:
		output_file_handle.write(",latency_cycles")
	output_file_handle.write("\n")

# Read every line of the input file
for line in input_file_handle:
	line=line.strip()
	inputs_values=line.split(",")
	if len(inputs_values)==0:
		continue
	elif len(inputs_values)==len(inputs):
		print ("Running simulation with inputs: "+line)

		# Remove the simbox file
		command="rm -f "+working_dir+"/simboxtemp.json"
		p = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, close_fds=True)
		p.wait()

		# Prepare the simbox general commands, like showing IOs, ticks, pc, disasm, etc.
		# These are useful for debugging and are disabled by default
		commands=[]
		commands.append("simbox -simbox-file "+working_dir+"/simboxtemp.json -add \"config:show_io_pre\"")
		commands.append("simbox -simbox-file "+working_dir+"/simboxtemp.json -suspend 0") 
		commands.append("simbox -simbox-file "+working_dir+"/simboxtemp.json -add \"config:show_io_post\"")
		commands.append("simbox -simbox-file "+working_dir+"/simboxtemp.json -suspend 1")
		commands.append("simbox -simbox-file "+working_dir+"/simboxtemp.json -add \"config:show_ticks\"")
		commands.append("simbox -simbox-file "+working_dir+"/simboxtemp.json -suspend 2")
		commands.append("simbox -simbox-file "+working_dir+"/simboxtemp.json -add \"config:show_pc\"")
		commands.append("simbox -simbox-file "+working_dir+"/simboxtemp.json -suspend 3")
		commands.append("simbox -simbox-file "+working_dir+"/simboxtemp.json -add \"config:show_disasm\"")
		commands.append("simbox -simbox-file "+working_dir+"/simboxtemp.json -suspend 4")

		for command in commands:
			p = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, close_fds=True)
			p.wait()
			if p.returncode!=0:
				print ("Error preparing simbox command: "+command)
				sys.exit(2)

		# Add the inputs to the simbox
		for i in range(len(inputs_values)):
			input_name=inputs[str(i)]
			input_value=inputs_values[i]
			command="simbox -simbox-file "+working_dir+"/simboxtemp.json -add \"absolute:0:set:"+input_name+":"+input_value+"\""
			p = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, close_fds=True)
			p.wait()
			if p.returncode!=0:
				print ("Error setting input "+input_name+" to "+input_value)
				sys.exit(2)

		# Prepare the outputs to be collected
		for output_name in outputs:
			if benchCore and outputs[output_name]=="o"+str(len(outputs)-1):
				command="simbox -simbox-file "+working_dir+"/simboxtemp.json -add \"onexit:show:"+outputs[output_name]+":unsigned\""
			else:
				command="simbox -simbox-file "+working_dir+"/simboxtemp.json -add \"onexit:show:"+outputs[output_name]+":"+data_type+"\""
			p = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, close_fds=True)
			p.wait()
			if p.returncode!=0:
				print ("Error getting output "+output_name)
				sys.exit(2)

		stopOnValidOf=len(outputs)-1

		# Run the simulation
		command="bondmachine -bondmachine-file "+working_dir+"/bondmachine.json -simbox-file "+working_dir+"/simboxtemp.json -sim-stop-on-valid-of "+str(stopOnValidOf)+" -sim -sim-interactions "+simulation_steps+" "+linear_data_range
		p = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, close_fds=True)
		p.wait()
		if p.returncode!=0:
			print ("Error running simulation")
			print (p.stderr.read().decode())
			sys.exit(2)

		outline=p.stdout.read().decode().strip()
		if omit_prefix:
			outline=outline.replace(prefix,"")

		# If there is an active benchCore mode, extract the latency cycles that is the last output
		if benchCore:
			parts=outline.split(' ')
			latency_cycles=parts[-1]
			outline=" ".join(parts[:-1])

		if isml:
			import numpy as np
			vals=np.asarray(outline.split(' '))
			vals=vals.astype(np.float32) # TODO: make this configurable
			index=np.argmax(vals)
			outline=outline.replace(" ",",")
			outline=outline+ "," + str(index)
		else:
			outline=outline.strip(',')
			outline=outline.replace(" ",",")

		# Write the output line eventually adding the latency cycles at the end
		output_file_handle.write(outline)
		if benchCore:
			output_file_handle.write(","+latency_cycles)
		output_file_handle.write("\n")

	else:
		print("Error: The input file has an invalid number of columns")

input_file_handle.close()
output_file_handle.close()
