
Double_t fitf(Double_t *v, Double_t *par)
{
	Double_t fitval = 0; 
	if(v[0] >= par[0]) {
		fitval = par[1] + par[2]*-1*TMath::Exp(-(v[0] - par[3])/par[4]) + par[5]*TMath::Exp(-(v[0] - par[3])/par[6]);
	}
	return fitval;
}

void plot(TString name, int evt)
{
	gStyle->SetPadGridX(true);
	gStyle->SetPadGridY(true);
	
	TFile* f = new TFile(name, "read");
	TTree* tree = (TTree*) f->Get("tree");
	
	// Make plot
	TCanvas* c1 = new TCanvas("c1","c1",900,900);
	c1->Divide(2,2);
	c1->cd(1);
	tree->Draw("T30[0] - T30[1]");
	c1->cd(2);
	tree->Draw("Pulse[1] : SampleTimes", Form("Evt == %i", evt),"goff");
	TGraph *g1 = new TGraph(999,tree->GetV2(),tree->GetV1());
	g1->SetMarkerColor(kRed);
	g1->SetMarkerStyle(8);
	g1->SetMarkerSize(0.5);
	tree->Draw("Pulse[0]*Ampl[1]/Ampl[0] : SampleTimes", Form("Evt == %i", evt),"same");
	TGraph *g2 = new TGraph(999,tree->GetV2(),tree->GetV1());
	g2->SetMarkerColor(kBlue);
	g2->SetMarkerStyle(8);
	g2->SetMarkerSize(0.5);
	TMultiGraph* mg = new TMultiGraph();
	mg->Add(g1);
	mg->Add(g2);
	mg->Draw("ap");
	
	// Retrieve values and print them on plot
	tree->Draw("Ampl[1]", Form("Evt == %i", evt),"goff");
	amplitude1 = *tree->GetV1();
	tree->Draw("Ampl[0]", Form("Evt == %i", evt),"goff");
	amplitude0 = *tree->GetV1();
	TLine* l = new TLine(0, amplitude1*0.3, 200, amplitude1*0.3);
	l->SetLineWidth(2);
	l->Draw("same");
	tree->Draw("T30[0] - T30[1]", Form("Evt == %i", evt),"goff");
	TLatex* la1 = new TLatex(110, amplitude1 * 0.6, Form("#Delta t = %f ns", *tree->GetV1()));
	la1->Draw();
	TLatex* la2 = new TLatex(110, amplitude1 * 0.8, Form("ampl[0]=%f", amplitude0));
	la2->SetTextColor(kBlue);
	la2->Draw();
	TLatex* la3 = new TLatex(110, amplitude1 * 0.7, Form("ampl[1]=%f", amplitude1));
	la3->SetTextColor(kRed);
	la3->Draw();
	
	// Make pulse plot with fit
	//TF1 *fPulse = new TF1("fPulse",fitf,0,200,7);
	//g1->Fit("fPulse", "", "", 20, 200);
	
	/*
	TF1* fPulse = new TF1("fPulse", "[0]*-1*TMath::Exp(-(x-[4])/[1]) + [2]*TMath::Exp(-(x - [4])/[3])", 0, 200);
	fPulse->SetLineColor(kBlack);
	fPulse->SetParameters(1, 4, 1, 40, 20);
	c1->cd(3);
	g1->Fit("fPulse", "", "", 20, 200);
	//fPulse->Draw();
	g1->Draw("ap");
	*/
	
	/*
	TF1* f1 = new TF1("fPulse","[2]*(1 + TMath::Erf((x-[0])/[1])) + [3]*TMath::Erf(-(x-[5])/[4])", 20, 200);
	f1->SetParameter(4, 40);
	f1->SetParameter(2, 1200);
	//f1->Draw();
	g1->Fit("fPulse", "R");
	*/
}